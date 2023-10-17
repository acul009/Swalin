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

	var pubkey *ecdsa.PublicKey = nil

	if conn.partner != nil {
		var ok bool
		pubkey, ok = conn.partner.PublicKey.(*ecdsa.PublicKey)

		if !ok {
			err := errors.New("invalid public key type")
			panic(err)
		}
	}

	return &RpcSession{
		Stream:     stream,
		ReadBuffer: make([]byte, 0, 1024),
		Connection: conn,
		Uuid:       uuid.New(),
		state:      RpcSessionCreated,
		partner:    pubkey,
	}

}

func (s *RpcSession) handleIncoming(commands *CommandCollection) {
	err := s.MutateState(RpcSessionCreated, RpcSessionOpen)
	if err != nil {
		log.Printf("session already open: %v", err)
		return
	}

	log.Printf("Session opened, reading request header...")

	header, sender, err := s.ReadRequestHeader()

	if sender == nil {
		s.Close()
		return
	}

	s.partner = sender

	if err != nil {
		s.WriteResponseHeader(SessionResponseHeader{
			Code: 500,
			Msg:  "error reading request header",
		})
		log.Printf("error reading header: %v", err)
		return
	}

	log.Printf("Header: %v", header)

	err = permissions.MayStartCommand(sender, header.Cmd)
	if err != nil {
		s.WriteResponseHeader(SessionResponseHeader{
			Code: 403,
			Msg:  "permission denied",
		})
		log.Printf("permission denied: %v", err)
		return
	}

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
		log.Printf("unknown command: %s", header.Cmd)
		return
	}

	cmd := handler()

	err = reEncode(header.Args, cmd)
	if err != nil {
		s.WriteResponseHeader(SessionResponseHeader{
			Code: 422,
			Msg:  "error unmarshalling command",
		})
		log.Printf("error unmarshalling command: %v", err)
		return
	}

	err = cmd.ExecuteServer(s)

	if err != nil {
		log.Printf("error executing command: %v", err)
		return
	}

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
		log.Printf("buffer: %s", string(s.ReadBuffer))
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
				return offset, fmt.Errorf("error seeking needle: %w", err)
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
	log.Printf("Reading message from unknown sender")

	message := &RpcMessage[P]{
		Payload: payload,
	}

	sender, err := pki.ReadAndUnmarshalAndVerify(s.Stream, message)
	if err != nil {
		return nil, fmt.Errorf("error reading message: %w", err)
	}

	myPub, err := pki.GetCurrentPublicKey()
	if err != nil {
		return nil, fmt.Errorf("error getting current key: %w", err)
	}

	err = message.Verify(s.Connection.nonceStorage, myPub)
	if err != nil {
		return nil, fmt.Errorf("error verifying message: %w", err)
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

	if !expectedSender.Equal(actualSender) {
		return fmt.Errorf("expected sender does not match actual sender")
	}

	return nil
}

func WriteMessage[P any](s *RpcSession, receiver *ecdsa.PublicKey, payload P) error {
	if err := s.EnsureState(RpcSessionOpen); err != nil {
		return fmt.Errorf("error ensuring state: %w", err)
	}

	pub, err := pki.GetCurrentPublicKey()
	if err != nil {
		return fmt.Errorf("error getting current key: %w", err)
	}

	key, err := pki.GetCurrentKey()
	if err != nil {
		return fmt.Errorf("error getting current key: %w", err)
	}

	if err != nil {
		return fmt.Errorf("error getting current key: %w", err)
	}

	message, err := newRpcMessage[P](receiver, payload)
	if err != nil {
		return fmt.Errorf("error creating message: %w", err)
	}

	data, err := pki.MarshalAndSign(message, key, pub)
	if err != nil {
		return fmt.Errorf("error marshalling message: %w", err)
	}

	log.Printf("Sending message: \n%s\n", string(data))

	n, err := s.Write(data)
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}
	if n != len(data) {
		return fmt.Errorf("error writing message: %w", io.ErrShortWrite)
	}

	return nil
}

func (s *RpcSession) WriteRequestHeader(receiver *ecdsa.PublicKey, header SessionRequestHeader) error {
	err := s.MutateState(RpcSessionCreated, RpcSessionOpen)
	if err != nil {
		return fmt.Errorf("error mutating state: %w", err)
	}

	err = WriteMessage[SessionRequestHeader](s, receiver, header)
	if err != nil {
		s.MutateState(RpcSessionOpen, RpcSessionClosed)
		return fmt.Errorf("error writing request header: %w", err)
	}

	err = s.MutateState(RpcSessionOpen, RpcSessionRequested)
	if err != nil {
		return fmt.Errorf("error mutating state: %w", err)
	}

	return nil
}

func (s *RpcSession) WriteResponseHeader(header SessionResponseHeader) error {
	err := s.MutateState(RpcSessionRequested, RpcSessionOpen)
	if err != nil {
		return fmt.Errorf("error mutating state: %w", err)
	}

	err = WriteMessage[SessionResponseHeader](s, s.partner, header)
	if err != nil {
		s.MutateState(s.state, RpcSessionClosed)
		return fmt.Errorf("error writing response header: %w", err)
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

func (s *RpcSession) SendCommand(receiver *ecdsa.PublicKey, cmd RpcCommand) error {

	args := make(map[string]interface{})
	err := reEncode(cmd, &args)
	if err != nil {
		return fmt.Errorf("error encoding command: %w", err)
	}

	header := SessionRequestHeader{
		Cmd:       cmd.GetKey(),
		Timestamp: time.Now().UnixMicro(),
		Args:      args,
	}

	err = s.WriteRequestHeader(receiver, header)
	if err != nil {
		return fmt.Errorf("error writing header to stream: %w", err)
	}

	response, err := s.readResponseHeader(s.partner)

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

	s.Connection.removeSession(s.Uuid)

	err := s.Stream.Close()
	return err
}

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
		return SessionRequestHeader{}, from, fmt.Errorf("error reading request header: %w", err)
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
