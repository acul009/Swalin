package rpc

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"rahnit-rmm/permissions"
	"rahnit-rmm/pki"
	"sync"
	"time"

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
	quic.Stream
	Connection *RpcConnection
	ReadBuffer []byte
	Uuid       uuid.UUID
	state      RpcSessionState
	mutex      sync.Mutex
	partner    *ecdsa.PublicKey
}

func NewRpcSession(stream quic.Stream, conn *RpcConnection) *RpcSession {
	return &RpcSession{
		Stream:     stream,
		ReadBuffer: make([]byte, 0, 1024),
		Connection: conn,
		Uuid:       uuid.New(),
		state:      RpcSessionCreated,
	}
}

func (s *RpcSession) handleIncoming(commands *CommandCollection) error {

	s.mutex.Lock()
	if s.state != RpcSessionCreated {
		s.mutex.Unlock()
		return fmt.Errorf("RPC session not created")
	}
	s.state = RpcSessionOpen
	s.mutex.Unlock()

	header, sender, err := s.ReadRequestHeader()
	if err != nil {
		log.Printf("error reading header: %v\n", err)
	}

	err = permissions.MayStartCommand(sender, header.Cmd)
	if err != nil {
		s.WriteResponseHeader(SessionResponseHeader{
			Code: 403,
			Msg:  "permission denied",
		})
		return fmt.Errorf("permission denied: %v", err)
	}

	s.partner = sender

	debug, err := json.Marshal(header)
	if err != nil {
		log.Printf("error marshalling header: %v\n", err)
	}

	fmt.Printf("\nHeader\n%v\n", string(debug))

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
		return fmt.Errorf("error unmarshalling command: %v", err)
	}

	err = cmd.ExecuteServer(s)

	if err != nil {
		log.Printf("error executing command: %v", err)
		return fmt.Errorf("error executing command: %v", err)
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
	return s.Stream.Write(p)
}

func (s *RpcSession) Read(p []byte) (n int, err error) {
	readRequestSize := len(p)

	readCounter := 0

	if len(s.ReadBuffer) > 0 {
		readCounter = copy(p, s.ReadBuffer)
		s.ReadBuffer = s.ReadBuffer[readCounter:]
	}

	readRequestSize -= readCounter

	if readRequestSize > 0 {
		readFromStreamCount, err := s.Stream.Read(p[readCounter:])
		readCounter += readFromStreamCount
		if err != nil {
			return readCounter, err
		}
	}

	return readCounter, nil
}

func (s *RpcSession) Peek(p []byte, offset int) (n int, err error) {
	readRequestSize := len(p)

	currentBufferSize := len(s.ReadBuffer)

	var eof error = nil

	if currentBufferSize < offset+readRequestSize {
		toRead := offset + readRequestSize - currentBufferSize
		readBuffer := make([]byte, toRead)
		n, err := s.Stream.Read(readBuffer)
		s.ReadBuffer = append(s.ReadBuffer, readBuffer[:n]...)
		if err != nil {
			if errors.Is(err, io.EOF) {
				eof = err
			} else {
				return n, err
			}
		}
	}

	if offset+readRequestSize > len(s.ReadBuffer) {
		readRequestSize = len(s.ReadBuffer) - offset
		if readRequestSize < 0 {
			return 0, eof
		}
	}

	n = copy(p, s.ReadBuffer[offset:offset+readRequestSize])

	return n, eof
}

func (s *RpcSession) Seek(needle []byte, bufferSize int, limit int) (offset int, err error) {
	needleLength := len(needle)
	readBuffer := make([]byte, bufferSize)

	offset = 0

	closed := false

	for offset+bufferSize < limit {
		readOffset := offset - needleLength
		if readOffset < 0 {
			readOffset = 0
		}

		n, err := s.Peek(readBuffer, readOffset)

		if err != nil {
			if err == io.EOF {
				closed = true
			} else {
				return offset, fmt.Errorf("error seeking needle: %v", err)
			}
		}

		i := findSubSlice(readBuffer[:n], needle)
		if i != -1 {
			return offset + i, nil
		}

		if closed {
			return i, fmt.Errorf("error seeking needle, EOF reached")
		}

		offset += n
	}

	return offset, fmt.Errorf("error seeking needle, limit reached")
}

func (s *RpcSession) ReadUntil(delimiter []byte, bufferSize int, limit int) ([]byte, error) {
	dataLength, err := s.Seek(delimiter, bufferSize, limit)
	if err != nil {
		return []byte{}, err
	}

	buffer := make([]byte, dataLength+len(delimiter))
	_, err = s.Read(buffer)
	if err != nil {
		return buffer, err
	}
	return buffer[:dataLength], nil
}

func readMessageFromUnknown[P any](s *RpcSession, payload P) (*ecdsa.PublicKey, error) {
	msgData, err := s.ReadUntil([]byte("\n"), 1024, 1024)
	if err != nil {
		return nil, fmt.Errorf("error reading message: %v", err)
	}

	message := rpcMessage[P]{}

	sender, err := pki.UnmarshalAndVerify(msgData, message)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling message: %v", err)
	}

	myPub, err := pki.GetCurrentPublicKey()
	if err != nil {
		return nil, fmt.Errorf("error getting current key: %v", err)
	}

	err = message.Verify(s.Connection.nonceStorage, myPub)
	if err != nil {
		return nil, fmt.Errorf("error verifying message: %v", err)
	}

	if err != nil {
		return nil, fmt.Errorf("error reading message: %v", err)
	}

	return sender, nil
}

func ReadMessage[P any](s *RpcSession, payload P, expectedSender *ecdsa.PublicKey) error {
	if expectedSender == nil {
		return fmt.Errorf("expected sender not found")
	}

	actualSender, err := readMessageFromUnknown[P](s, payload)
	if err != nil {
		return err
	}

	expected, err := pki.EncodePubToString(expectedSender)
	if err != nil {
		return fmt.Errorf("failed to encode public key: %v", err)
	}

	actual, err := pki.EncodePubToString(actualSender)
	if err != nil {
		return fmt.Errorf("failed to encode public key: %v", err)
	}

	if expected != actual {
		return fmt.Errorf("expected sender does not match actual sender")
	}

	return nil
}

func WriteMessage[P any](s *RpcSession, payload P) error {
	if err := s.EnsureState(RpcSessionOpen); err != nil {
		return fmt.Errorf("error ensuring state: %v", err)
	}

	pub, err := pki.GetCurrentPublicKey()
	if err != nil {
		return fmt.Errorf("error getting current key: %v", err)
	}

	key, err := pki.GetCurrentKey()
	if err != nil {
		return fmt.Errorf("error getting current key: %v", err)
	}

	if err != nil {
		return fmt.Errorf("error getting current key: %v", err)
	}

	data, err := pki.MarshalAndSign(payload, key, pub)
	if err != nil {
		return fmt.Errorf("error marshalling message: %v", err)
	}

	toWrite := append(data, messageStop...)

	_, err = s.Write(toWrite)
	if err != nil {
		return fmt.Errorf("error writing message: %v", err)
	}

	return nil
}

func (s *RpcSession) WriteRequestHeader(header SessionRequestHeader) error {
	err := s.MutateState(RpcSessionCreated, RpcSessionOpen)
	if err != nil {
		return fmt.Errorf("error mutating state: %v", err)
	}

	err = WriteMessage[SessionRequestHeader](s, header)
	if err != nil {
		s.MutateState(RpcSessionOpen, RpcSessionClosed)
		return fmt.Errorf("error writing request header: %v", err)
	}

	err = s.MutateState(RpcSessionOpen, RpcSessionRequested)
	if err != nil {
		return fmt.Errorf("error mutating state: %v", err)
	}

	return nil
}

func (s *RpcSession) WriteResponseHeader(header SessionResponseHeader) error {
	err := s.MutateState(RpcSessionRequested, RpcSessionOpen)
	if err != nil {
		return fmt.Errorf("error mutating state: %v", err)
	}

	err = WriteMessage[SessionResponseHeader](s, header)
	if err != nil {
		s.MutateState(s.state, RpcSessionClosed)
		return fmt.Errorf("error writing response header: %v", err)
	}

	return nil
}

func (s *RpcSession) MutateState(from RpcSessionState, to RpcSessionState) error {
	s.mutex.Lock()
	if s.state != from {
		s.mutex.Unlock()
		return fmt.Errorf("RPC session not in state %v", from)
	}
	s.state = to
	s.mutex.Unlock()
	return nil
}

func (s *RpcSession) EnsureState(state RpcSessionState) error {
	s.mutex.Lock()
	if s.state != state {
		s.mutex.Unlock()
		return fmt.Errorf("RPC session not in state %v", state)
	}
	s.mutex.Unlock()
	return nil
}

func (s *RpcSession) SendCommand(cmd RpcCommand) error {

	args := make(map[string]interface{})
	err := reEncode(cmd, &args)
	if err != nil {
		return fmt.Errorf("error encoding command: %v", err)
	}

	header := SessionRequestHeader{
		Cmd:       cmd.GetKey(),
		Timestamp: time.Now().UnixMicro(),
		Args:      args,
	}

	err = s.WriteRequestHeader(header)
	if err != nil {
		return fmt.Errorf("error writing header to stream: %v", err)
	}

	response, err := s.readResponseHeader(s.partner)

	fmt.Printf("Response Header:\n%v\n", response)

	if err != nil {
		return fmt.Errorf("error reading response header: %v", err)
	}

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

	s.Connection.removeSession(s.Uuid)

	err := s.Stream.Close()
	return err
}

var messageStop = []byte("\n")
var headerDelimiter = "|"

type SessionRequestHeader struct {
	Cmd       string                 `json:"cmd"`
	Timestamp int64                  `json:"timestamp"`
	Args      map[string]interface{} `json:"args"`
}

type SessionResponseHeader struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Info interface{} `json:"info"`
}

func (s *RpcSession) ReadRequestHeader() (SessionRequestHeader, *ecdsa.PublicKey, error) {
	header := SessionRequestHeader{}
	from, err := readMessageFromUnknown[SessionRequestHeader](s, header)
	if err != nil {
		return SessionRequestHeader{}, nil, fmt.Errorf("error reading request header: %v", err)
	}

	return header, from, nil
}

func (s *RpcSession) readResponseHeader(expectedSender *ecdsa.PublicKey) (SessionResponseHeader, error) {

	s.mutex.Lock()
	if s.state != RpcSessionRequested {
		s.mutex.Unlock()
		return SessionResponseHeader{}, fmt.Errorf("RPC connection not yet requested")
	}
	s.mutex.Unlock()

	header := SessionResponseHeader{}
	err := ReadMessage[SessionResponseHeader](s, header, expectedSender)
	if err != nil {
		return SessionResponseHeader{}, fmt.Errorf("error reading response header: %v", err)
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

func findSubSlice(bigSlice, subSlice []byte) int {
	for i := 0; i < len(bigSlice)-len(subSlice)+1; i++ {
		if bytesEqual(bigSlice[i:i+len(subSlice)], subSlice) {
			return i
		}
	}
	return -1 // Return -1 if subSlice is not found in bigSlice
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func reEncode(from interface{}, to interface{}) error {
	data, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, to)
}
