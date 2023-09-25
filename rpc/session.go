package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/quic-go/quic-go"
)

func NewRpcSession(stream quic.Stream, conn *NodeConnection) *RpcSession {
	return &RpcSession{
		Stream:       stream,
		ReadBuffer:   make([]byte, 0, 1024),
		Connection:   conn,
		ReadyToWrite: true,
	}
}

type RpcSession struct {
	quic.Stream
	Connection   *NodeConnection
	ReadBuffer   []byte
	ReadyToWrite bool
}

func (s *RpcSession) Write(p []byte) (n int, err error) {
	if !s.ReadyToWrite {
		return 0, fmt.Errorf("Not ready to write")
	}
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
				return offset, fmt.Errorf("Error seeking needle: %v", err)
			}
		}

		i := findSubSlice(readBuffer[:n], needle)
		if i != -1 {
			return offset + i, nil
		}

		if closed {
			return i, fmt.Errorf("Error seeking needle, EOF reached")
		}

		offset += n
	}

	return offset, fmt.Errorf("Error seeking needle, limit reached")
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
	return s.writeRawHeader(header)
}

func (s *RpcSession) WriteResponseHeader(header SessionResponseHeader) (n int, err error) {
	if header.Code == 200 {
		s.ReadyToWrite = true
	}
	return s.writeRawHeader(header)
}

func (s *RpcSession) writeRawHeader(header interface{}) (n int, err error) {
	headerData, err := json.Marshal(header)
	payload := append(headerData, headerStop...)
	n, err = s.Stream.Write(payload)
	if err != nil {
		fmt.Printf("Error writing header to stream: %v\n", err)
		return n, fmt.Errorf("Error writing header to stream: %v", err)
	}
	fmt.Printf("Header written: %v\n", string(payload))
	return n, nil
}

func (s *RpcSession) SendCommand(cmd RpcCommand) error {

	args := make(map[string]interface{})
	err := reEncode(cmd, &args)
	if err != nil {
		return fmt.Errorf("Error encoding command: %v", err)
	}
	header := SessionRequestHeader{
		Cmd:       cmd.GetKey(),
		Timestamp: time.Now().UnixMicro(),
		Args:      args,
	}
	_, err = s.WriteRequestHeader(header)
	if err != nil {
		return fmt.Errorf("Error writing header to stream: %v", err)
	}

	response, err := ReadResponseHeader(s)

	fmt.Printf("Response Header:\n%v\n", response)

	if err != nil {
		return fmt.Errorf("Error reading response header: %v", err)
	}

	if response.Code != 200 {
		return fmt.Errorf("Error sending command: %v", response.Msg)
	}

	return cmd.ExecuteClient(s)
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
