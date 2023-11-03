package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"rahnit-rmm/permissions"
	"rahnit-rmm/pki"
	"sync"

	"github.com/google/uuid"
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
	uuid        uuid.UUID
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
		stream:      stream,
		ctx:         stream.Context(),
		connection:  conn,
		uuid:        uuid.New(),
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
		return fmt.Errorf("error executing command: %w", err)
	}

	return nil
}

func (s *RpcSession) Write(p []byte) (n int, err error) {
	s.mutex.Lock()
	if s.state != RpcSessionOpen {
		s.mutex.Unlock()
		return 0, fmt.Errorf("RPC session not open")
	}
	s.mutex.Unlock()
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

	sender, err := pki.ReadAndUnmarshalAndVerify(s.stream, message)
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
	if err := s.ensureState(RpcSessionOpen); err != nil {
		return fmt.Errorf("error ensuring state: %w", err)
	}

	if s.partner == nil {
		return fmt.Errorf("can't address message to unknown sender: session has no partner specified")
	}

	message, err := newRpcMessage[P](s.partner, payload)
	if err != nil {
		return fmt.Errorf("error creating message: %w", err)
	}

	data, err := pki.MarshalAndSign(message, s.credentials)
	if err != nil {
		return fmt.Errorf("error marshalling message: %w", err)
	}

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
	if s.state != from {
		s.mutex.Unlock()
		return fmt.Errorf("RPC session not in state %v", from)
	}
	s.state = to
	s.mutex.Unlock()
	return nil
}

func (s *RpcSession) ensureState(state RpcSessionState) error {
	s.mutex.Lock()
	if s.state != state {
		s.mutex.Unlock()
		return fmt.Errorf("RPC session not in state %v", state)
	}
	s.mutex.Unlock()
	return nil
}

func (s *RpcSession) sendCommand(cmd RpcCommand) error {

	args := make(map[string]interface{})
	err := reEncode(cmd, &args)
	if err != nil {
		return fmt.Errorf("error encoding command: %w", err)
	}

	header := SessionRequestHeader{
		Cmd:  cmd.GetKey(),
		Args: args,
	}

	err = s.writeRequestHeader(header)
	if err != nil {
		return fmt.Errorf("error writing header to stream: %w", err)
	}

	response, err := s.readResponseHeader()

	if err != nil {
		return fmt.Errorf("error reading response header: %w", err)
	}

	fmt.Printf("Response Header:\n%v\n", response)

	if response.Code != 200 {
		return fmt.Errorf("error sending command: %v", response.Msg)
	}

	return cmd.ExecuteClient(s)
}

func (s *RpcSession) Close() error {
	s.mutex.Lock()
	if s.state != RpcSessionOpen {
		s.mutex.Unlock()
		return fmt.Errorf("RPC session not open")
	}
	s.state = RpcSessionClosed
	s.mutex.Unlock()

	s.connection.removeSession(s.uuid)

	err := s.stream.Close()
	return err
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

	s.mutex.Lock()
	if s.state != RpcSessionRequested {
		s.mutex.Unlock()
		return SessionResponseHeader{}, fmt.Errorf("RPC connection not yet requested")
	}
	s.mutex.Unlock()

	header := SessionResponseHeader{}
	err := ReadMessage[*SessionResponseHeader](s, &header)
	if err != nil {
		return SessionResponseHeader{}, fmt.Errorf("error reading response header: %w", err)
	}

	s.mutex.Lock()
	if header.Code >= 200 || header.Code <= 299 {
		s.state = RpcSessionOpen
	} else {
		s.state = RpcSessionClosed
	}
	s.mutex.Unlock()

	return header, nil
}

func reEncode(from interface{}, to interface{}) error {
	data, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, to)
}
