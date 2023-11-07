package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"rahnit-rmm/permissions"
	"rahnit-rmm/pki"
	"rahnit-rmm/util"
	"strings"
	"sync"

	"github.com/quic-go/quic-go"
)

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
	partner     *pki.PublicKey
	credentials pki.Credentials
}

func newRpcSession(stream quic.Stream, conn *RpcConnection) *RpcSession {

	var pubkey *pki.PublicKey = nil

	if conn.partner != nil {
		pubkey = conn.partner.GetPublicKey()
	}

	return &RpcSession{
		stream:      wrapQuicStream(stream),
		ctx:         stream.Context(),
		connection:  conn,
		id:          stream.StreamID(),
		state:       RpcSessionCreated,
		mutex:       sync.Mutex{},
		partner:     pubkey,
		credentials: conn.credentials,
	}

}

func (s *RpcSession) handleIncoming(commands *CommandCollection) error {
	defer s.Close()
	log.Printf("handling incoming session...")
	err := s.ensureState(RpcSessionCreated)
	if err != nil {
		return fmt.Errorf("error ensuring state: %w", err)
	}

	header, sender, err := s.readRequestHeader()

	if err != nil {
		s.WriteResponseHeader(SessionResponseHeader{
			Code: 500,
			Msg:  "error reading request header",
		})
		return fmt.Errorf("error reading request header: %w", err)
	}

	_, err = s.connection.verifier.VerifyPublicKey(sender)
	if err != nil {
		s.WriteResponseHeader(SessionResponseHeader{
			Code: 401,
			Msg:  "error verifying public key",
		})
		return fmt.Errorf("error verifying public key: %w", err)
	}

	s.partner = sender

	log.Printf("Header: %+v", header)

	err = permissions.MayStartCommand(sender, header.Cmd)
	if err != nil {
		s.WriteResponseHeader(SessionResponseHeader{
			Code: 403,
			Msg:  "permission denied",
		})
		return fmt.Errorf("permission denied: %w", err)
	}

	handler, ok := commands.Get(header.Cmd)
	if !ok {
		s.WriteResponseHeader(SessionResponseHeader{
			Code: 404,
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

func readMessageFromUnknown[P any](s *RpcSession, payload P) (*pki.PublicKey, error) {

	message := &RpcMessage[P]{
		Payload: payload,
	}

	sender, err := pki.ReadAndUnmarshalAndVerify(s.stream, message, false)
	if err != nil {
		return nil, fmt.Errorf("error reading message: %w", err)
	}

	myPub, err := s.credentials.GetPublicKey()
	if err != nil {
		return nil, fmt.Errorf("error getting current public key: %w", err)
	}

	err = message.Verify(s.connection.nonceStorage, myPub)
	if err != nil {
		return nil, fmt.Errorf("error verifying message: %w", err)
	}

	return sender, nil
}

func ReadMessage[P any](s *RpcSession, payload P) error {
	if s.partner == nil {
		return fmt.Errorf("can't read message from unknown sender: session has no partner specified")
	}

	actualSender, err := readMessageFromUnknown[P](s, payload)
	if err != nil {
		return err
	}

	if !actualSender.Equal(s.partner) {
		return fmt.Errorf("expected sender does not match actual sender")
	}

	return nil
}

func WriteMessage[P any](s *RpcSession, payload P) error {
	if s.partner == nil {
		return fmt.Errorf("can't address message to unknown sender: session has no partner specified")
	}

	// log.Printf("creating message...")

	message, err := newRpcMessage[P](s.partner, payload)
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

func (s *RpcSession) writeRequestHeader(header SessionRequestHeader) error {
	err := s.mutateState(RpcSessionCreated, RpcSessionOpen)
	if err != nil {
		return fmt.Errorf("error mutating state: %w", err)
	}

	err = WriteMessage[SessionRequestHeader](s, header)
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

	args := make(map[string]interface{})
	err := reEncode(cmd, &args)
	if err != nil {
		return nil, fmt.Errorf("error encoding command: %w", err)
	}

	header := SessionRequestHeader{
		Cmd:  cmd.GetKey(),
		Args: args,
	}

	err = s.writeRequestHeader(header)
	if err != nil {
		return nil, fmt.Errorf("error writing header to stream: %w", err)
	}

	response, err := s.readResponseHeader()

	if err != nil {
		return nil, fmt.Errorf("error reading response header: %w", err)
	}

	fmt.Printf("Response Header:\n%v\n", response)

	if response.Code != 200 {
		return nil, fmt.Errorf("error sending command: %v", response.Msg)
	}

	running := &runningCommand{
		session: s,
		errChan: make(chan error),
	}

	go func() {
		defer func() {
			err := s.Close()
			if err != nil {
				panic(err)
			}
		}()
		running.errChan <- cmd.ExecuteClient(s)
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

type SessionRequestHeader struct {
	Cmd  string                 `json:"cmd"`
	Args map[string]interface{} `json:"args"`
}

type SessionResponseHeader struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Info interface{} `json:"info"`
}

func (s *RpcSession) readRequestHeader() (SessionRequestHeader, *pki.PublicKey, error) {
	header := SessionRequestHeader{}
	from, err := readMessageFromUnknown[*SessionRequestHeader](s, &header)
	if err != nil {
		return SessionRequestHeader{}, from, fmt.Errorf("error reading request header: %w", err)
	}

	err = s.mutateState(RpcSessionCreated, RpcSessionRequested)
	if err != nil {
		return SessionRequestHeader{}, from, fmt.Errorf("error setting session state: %w", err)
	}

	return header, from, nil
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
