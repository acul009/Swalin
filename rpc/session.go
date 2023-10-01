package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
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

	header, err := s.ReadRequestHeader()
	if err != nil {
		log.Printf("error reading header: %v\n", err)
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

func (s *RpcSession) WriteRequestHeader(header SessionRequestHeader) (n int, err error) {
	s.mutex.Lock()
	if s.state != RpcSessionCreated {
		s.mutex.Unlock()
		return 0, fmt.Errorf("RPC session not already in use")
	}
	s.state = RpcSessionRequested
	s.mutex.Unlock()
	return s.writeRawHeader(header)
}

func (s *RpcSession) WriteResponseHeader(header SessionResponseHeader) (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.state != RpcSessionOpen {
		return 0, fmt.Errorf("RPC connection not open")
	}

	n, err := s.writeRawHeader(header)
	s.state = RpcSessionClosed
	if err != nil {
		return n, fmt.Errorf("error writing header to stream: %v", err)
	}

	return n, nil
}

func (s *RpcSession) writeRawHeader(header interface{}) (n int, err error) {
	headerData, err := json.Marshal(header)
	if err != nil {
		return n, fmt.Errorf("error marshalling header: %v", err)
	}
	payload := append(headerData, headerStop...)
	n, err = s.Stream.Write(payload)
	if err != nil {
		fmt.Printf("error writing header to stream: %v\n", err)
		return n, fmt.Errorf("error writing header to stream: %v", err)
	}
	fmt.Printf("Header written: %v\n", string(payload))
	return n, nil
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
	_, err = s.WriteRequestHeader(header)
	if err != nil {
		return fmt.Errorf("error writing header to stream: %v", err)
	}

	response, err := s.readResponseHeader()

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
	err := s.Stream.Close()
	return err
}

var headerStop = []byte("\n")
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

func (s *RpcSession) ReadRequestHeader() (SessionRequestHeader, error) {
	headerData, err := s.ReadUntil(headerStop, 1024, 65536)
	if err != nil {
		return SessionRequestHeader{}, err
	}

	header := string(headerData)

	headerParts := strings.Split(header, headerDelimiter)

	decodedHeader := SessionRequestHeader{}
	err = json.Unmarshal([]byte(headerParts[0]), &decodedHeader)
	if err != nil {
		return SessionRequestHeader{}, err
	}

	return decodedHeader, nil
}

func (s *RpcSession) readResponseHeader() (SessionResponseHeader, error) {

	s.mutex.Lock()
	if s.state != RpcSessionRequested {
		s.mutex.Unlock()
		return SessionResponseHeader{}, fmt.Errorf("RPC connection not yet requested")
	}
	s.mutex.Unlock()

	headerData, err := s.ReadUntil(headerStop, 1024, 65536)
	if err != nil {
		return SessionResponseHeader{}, err
	}

	header := string(headerData)

	headerParts := strings.Split(header, headerDelimiter)

	decodedHeader := SessionResponseHeader{}
	err = json.Unmarshal([]byte(headerParts[0]), &decodedHeader)
	if err != nil {
		return SessionResponseHeader{}, err
	}

	s.mutex.Lock()
	if decodedHeader.Code >= 200 || decodedHeader.Code <= 299 {
		s.state = RpcSessionOpen
	} else {
		s.state = RpcSessionClosed
	}
	s.mutex.Unlock()

	return decodedHeader, nil
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
