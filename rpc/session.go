package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/util"

	"github.com/quic-go/quic-go"
)

type SessionError struct {
	code int
	msg  string
}

func (e *SessionError) Error() string {
	return fmt.Sprintf("%d: %s", e.code, e.msg)
}

func (e *SessionError) Is(target error) bool {
	_, ok := target.(*SessionError)
	return ok
}

func (e *SessionError) As(target interface{}) bool {
	if err, ok := target.(*SessionError); ok {
		*err = *e
		return true
	}
	return false
}

func (e *SessionError) Code() int {
	return e.code
}

type RpcSessionState int16

const (
	RpcSessionCreated RpcSessionState = iota
	RpcSessionRequested
	RpcSessionOpen
	RpcSessionClosed
)

type RpcSession struct {
	stream      io.ReadWriteCloser
	ctx         context.Context
	connection  *RpcConnection
	id          quic.StreamID
	state       RpcSessionState
	mutex       sync.Mutex
	partnerKey  *pki.PublicKey
	partner     *pki.Certificate
	credentials pki.Credentials
}

func newRpcSession(stream quic.Stream, conn *RpcConnection) *RpcSession {

	var pubkey *pki.PublicKey = nil

	if conn.partner != nil {
		pubkey = conn.partner.PublicKey()
	}

	return &RpcSession{
		stream:      wrapQuicStream(stream),
		ctx:         stream.Context(),
		connection:  conn,
		id:          stream.StreamID(),
		state:       RpcSessionCreated,
		mutex:       sync.Mutex{},
		partnerKey:  pubkey,
		credentials: conn.credentials,
	}

}

func (s *RpcSession) handleIncoming(commands *CommandCollection) error {
	defer s.Close()
	log.Printf("handling incoming session...")

	err := s.mutateState(RpcSessionCreated, RpcSessionOpen)
	if err != nil {
		return fmt.Errorf("error ensuring state: %w", err)
	}

	err = receivePartnerKey(s)
	if err != nil {
		return fmt.Errorf("error receiving partner key: %w", err)
	}

	s.mutateState(RpcSessionOpen, RpcSessionCreated)

	chain, err := s.Verifier().VerifyPublicKey(s.partnerKey)
	if err != nil {
		s.mutateState(RpcSessionCreated, RpcSessionRequested)
		s.WriteResponseHeader(SessionResponseHeader{
			Code: 401,
			Msg:  "error verifying public key",
		})
		return fmt.Errorf("error verifying public key: %w", err)
	}

	s.partner = chain[0]

	header, err := s.readRequestHeader()

	if err != nil {
		s.WriteResponseHeader(SessionResponseHeader{
			Code: 500,
			Msg:  "error reading request header",
		})
		return fmt.Errorf("error reading request header: %w", err)
	}

	log.Printf("Header: %+v", header)

	// TODO: Check if the command is allowed (permissions)

	handler, ok := commands.Get(header.Cmd)
	if !ok {
		s.WriteResponseHeader(SessionResponseHeader{
			Code: 405,
			Msg:  "command not found",
		})
		return fmt.Errorf("unknown command: %s", header.Cmd)
	}

	cmd := handler()

	err = reEncode(header.Args, cmd)
	if err != nil {
		s.WriteResponseHeader(SessionResponseHeader{
			Code: 422,
			Msg:  "error unmarshalling command",
		})
		return fmt.Errorf("error unmarshalling command: %w", err)
	}

	err = cmd.ExecuteServer(s)

	if err != nil {
		if s.ensureState(RpcSessionClosed) == nil {
			s.WriteResponseHeader(SessionResponseHeader{
				Code: 500,
				Msg:  "error executing command",
			})
			return fmt.Errorf("error executing command: %w", err)
		}
		return fmt.Errorf("error executing command: %w", err)
	} else {
		if s.ensureState(RpcSessionClosed) == nil {
			s.WriteResponseHeader(SessionResponseHeader{
				Code: 500,
				Msg:  "error executing command",
			})
			return fmt.Errorf("error executing command: %w", err)
		}
	}

	return nil
}

func (s *RpcSession) Write(p []byte) (n int, err error) {
	err = s.ensureState(RpcSessionOpen)
	if err != nil {
		return 0, fmt.Errorf("error ensuring state: %w", err)
	}
	// log.Printf("Writing raw data to stream...")
	return s.stream.Write(p)
}

func (s *RpcSession) Read(p []byte) (n int, err error) {
	return s.stream.Read(p)
}

func (s *RpcSession) Context() context.Context {
	return s.ctx
}

func ReadMessage[P any](s *RpcSession, payload P) error {
	if s.partnerKey == nil {
		return fmt.Errorf("can't read message from unknown sender: session has no partner specified")
	}

	message := &RpcMessage[P]{
		Payload: payload,
	}

	err := pki.ReadAndUnmarshalAndVerify(s.stream, message, s.partnerKey, false)
	if err != nil {
		return fmt.Errorf("error reading message: %w", err)
	}

	err = message.Verify(s.connection.nonceStorage, s.credentials.PublicKey())
	if err != nil {
		return fmt.Errorf("error verifying message: %w", err)
	}

	return nil
}

func WriteMessage[P any](s *RpcSession, payload P) error {
	if s.partnerKey == nil {
		return fmt.Errorf("can't address message to unknown sender: session has no partner specified")
	}

	// log.Printf("creating message...")

	message, err := newRpcMessage[P](s.partnerKey, payload)
	if err != nil {
		return fmt.Errorf("error creating message: %w", err)
	}

	// log.Printf("marshalling message...")

	data, err := pki.MarshalAndSign(message, s.credentials)
	if err != nil {
		return fmt.Errorf("error marshalling message: %w", err)
	}

	// log.Printf("writing message...")

	n, err := s.Write(data)
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}
	if n != len(data) {
		return fmt.Errorf("error writing message: %w", io.ErrShortWrite)
	}

	return nil
}

func (s *RpcSession) writeRequestHeader(header sessionRequestHeader) error {
	err := s.mutateState(RpcSessionCreated, RpcSessionOpen)
	if err != nil {
		return fmt.Errorf("error mutating state: %w", err)
	}

	err = WriteMessage[sessionRequestHeader](s, header)
	if err != nil {
		s.mutateState(RpcSessionOpen, RpcSessionClosed)
		return fmt.Errorf("error writing request header: %w", err)
	}

	err = s.mutateState(RpcSessionOpen, RpcSessionRequested)
	if err != nil {
		return fmt.Errorf("error mutating state: %w", err)
	}

	return nil
}

func (s *RpcSession) WriteResponseHeader(header SessionResponseHeader) error {
	err := s.mutateState(RpcSessionRequested, RpcSessionOpen)
	if err != nil {
		return fmt.Errorf("error mutating state: %w", err)
	}

	err = WriteMessage[SessionResponseHeader](s, header)
	if err != nil {
		stateErr := s.mutateState(s.state, RpcSessionClosed)
		if stateErr != nil {
			panic(stateErr)
		}
		return fmt.Errorf("error writing response header: %w", err)
	}

	return nil
}

func (s *RpcSession) mutateState(from RpcSessionState, to RpcSessionState) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.state != from {
		return fmt.Errorf("RPC session not in state %v", from)
	}
	s.state = to
	return nil
}

func (s *RpcSession) ensureState(state RpcSessionState) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.state != state {
		return fmt.Errorf("RPC session not in state %v", state)
	}
	return nil
}

func (s *RpcSession) sendCommand(cmd RpcCommand) (util.AsyncAction, error) {
	err := s.mutateState(RpcSessionCreated, RpcSessionOpen)
	if err != nil {
		return nil, fmt.Errorf("error mutating state: %w", err)
	}

	err = sendMyKey(s)
	if err != nil {
		s.Close()
		return nil, fmt.Errorf("error sending my public key: %w", err)
	}

	err = s.mutateState(RpcSessionOpen, RpcSessionCreated)
	if err != nil {
		return nil, fmt.Errorf("error mutating state: %w", err)
	}

	args := make(map[string]interface{})
	err = reEncode(cmd, &args)
	if err != nil {
		s.Close()
		return nil, fmt.Errorf("error encoding command: %w", err)
	}

	header := sessionRequestHeader{
		Cmd:  cmd.GetKey(),
		Args: args,
	}

	err = s.writeRequestHeader(header)
	if err != nil {
		s.Close()
		return nil, fmt.Errorf("error writing header to stream: %w", err)
	}

	response, err := s.readResponseHeader()
	if err != nil {
		s.Close()
		return nil, fmt.Errorf("error reading response header: %w", err)
	}

	log.Printf("Response Header:\n%v\n", response)

	if response.Code != 200 {
		s.Close()
		return nil, fmt.Errorf("error sending command: %w", &SessionError{
			code: response.Code,
			msg:  response.Msg,
		})
	}

	log.Printf("Command sent successfully: %s", cmd.GetKey())

	running := &runningCommand{
		session: s,
		errChan: make(chan error),
	}

	go func() {
		running.errChan <- cmd.ExecuteClient(s)
		err := s.Close()
		if err != nil {
			panic(err)
		}
	}()

	return running, nil
}

type runningCommand struct {
	session    *RpcSession
	errChan    chan error
	forceClose bool
}

func (r *runningCommand) Close() error {
	r.forceClose = true
	err := r.session.Close()
	if err != nil {
		return fmt.Errorf("error closing session: %w", err)
	}

	return nil
}

func (r *runningCommand) Wait() error {
	err := <-r.errChan
	if err != nil {
		// cancel errors are expected, since we might be force closing
		streamErr := &quic.StreamError{}
		if r.forceClose && errors.Is(err, streamErr) {
			return nil
		}
	}

	return nil
}

func (s *RpcSession) Close() error {
	log.Printf("closing session...")
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.state == RpcSessionClosed {
		return nil
	}
	s.state = RpcSessionClosed

	s.connection.removeSession(s.id)

	err := s.stream.Close()
	if err != nil {
		return fmt.Errorf("error closing underlying readwriter: %w", err)
	}

	return nil
}

type sessionRequestHeader struct {
	Cmd  string                 `json:"cmd"`
	Args map[string]interface{} `json:"args"`
}

type SessionResponseHeader struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Info interface{} `json:"info"`
}

func (s *RpcSession) readRequestHeader() (sessionRequestHeader, error) {
	header := sessionRequestHeader{}
	err := ReadMessage[*sessionRequestHeader](s, &header)
	if err != nil {
		return sessionRequestHeader{}, fmt.Errorf("error reading request header: %w", err)
	}

	err = s.mutateState(RpcSessionCreated, RpcSessionRequested)
	if err != nil {
		return sessionRequestHeader{}, fmt.Errorf("error setting session state: %w", err)
	}

	return header, nil
}

func (s *RpcSession) readResponseHeader() (SessionResponseHeader, error) {
	err := s.ensureState(RpcSessionRequested)
	if err != nil {
		return SessionResponseHeader{}, fmt.Errorf("error ensuring state: %w", err)
	}

	header := SessionResponseHeader{}
	err = ReadMessage[*SessionResponseHeader](s, &header)
	if err != nil {
		return SessionResponseHeader{}, fmt.Errorf("error reading response header: %w", err)
	}

	if header.Code >= 200 || header.Code <= 299 {
		err = s.mutateState(RpcSessionRequested, RpcSessionOpen)
		if err != nil {
			return SessionResponseHeader{}, fmt.Errorf("error mutating state: %w", err)
		}
	} else {
		err = s.mutateState(RpcSessionRequested, RpcSessionClosed)
		if err != nil {
			return SessionResponseHeader{}, fmt.Errorf("error mutating state: %w", err)
		}
	}

	return header, nil
}

func reEncode(from interface{}, to interface{}) error {
	data, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, to)
}

type streamWrapper struct {
	quic.Stream
}

func wrapQuicStream(stream quic.Stream) *streamWrapper {
	return &streamWrapper{stream}
}

func (s *streamWrapper) Close() error {
	log.Printf("Closing stream wrapper")
	err := s.Stream.Close()
	if err != nil {
		if !strings.HasPrefix(err.Error(), "close called for canceled stream") {
			return fmt.Errorf("error closing stream wrapper: %w", err)
		}
	}
	s.Stream.CancelRead(200)
	return nil
}

func (s *RpcSession) Verifier() pki.Verifier {
	return s.connection.verifier
}

func (s *RpcSession) Partner() *pki.Certificate {
	return s.partner
}
