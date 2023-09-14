package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/quic-go/quic-go"
)

func ServeSession(conn quic.Connection, commands *CommandCollection) {
	fmt.Println("Connection accepted, serving RPC")
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Printf("Error accepting QUIC stream: %v", err)
			return
		}
		go handleSession(NewRpcSession(stream))
	}

}

func NewRpcSession(conn quic.Stream) *RpcSession {
	return &RpcSession{
		Stream:     conn,
		ReadBuffer: make([]byte, 0, 1024),
	}
}

type RpcSession struct {
	quic.Stream
	ReadBuffer []byte
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

	if currentBufferSize < offset+readRequestSize {
		toRead := offset + readRequestSize - currentBufferSize
		readBuffer := make([]byte, toRead)
		n, err := s.Stream.Read(readBuffer)
		s.ReadBuffer = append(s.ReadBuffer, readBuffer[:n]...)
		if err != nil {
			if err == io.EOF {

			} else {
				return n, err
			}
		}
	}

	n = copy(p, s.ReadBuffer[offset:offset+readRequestSize])

	return n, nil
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

		offset += bufferSize
	}

	return offset, fmt.Errorf("Error seeking needle, limit reached")
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

func handleSession(session *RpcSession) {
	header, err := ReadHeader(session)
	if err != nil {
		log.Printf("Error reading header: %v\n", err)
	}

	debug, err := json.Marshal(header)
	fmt.Printf("\nHeader\n%v\n", string(debug))
}

type SessionRequestHeader struct {
	Type      string                 `json:"type"`
	Timestamp int64                  `json:"timestamp"`
	Args      map[string]interface{} `json:"args"`
}

type SessionResponseHeader struct {
	Code int `json:"code"`
}

var headerStop = []byte("\n")
var headerDelimiter = "|"

func ReadHeader(session *RpcSession) (SessionRequestHeader, error) {
	headerLength, err := session.Seek(headerStop, 1024, 65536)
	if err != nil {
		return SessionRequestHeader{}, err
	}

	headerData := make([]byte, headerLength)
	_, err = session.Read(headerData)
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
