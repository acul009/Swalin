package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/quic-go/quic-go"
)

func ServeSession(conn quic.Connection, commands *CommandCollection) {
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Printf("Error accepting QUIC stream: %v", err)
			continue
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

func (s *RpcSession) Peek(p []byte) (n int, err error) {
	readRequestSize := len(p)

	readCounter := 0

	if len(s.ReadBuffer) > 0 {
		readCounter = copy(p, s.ReadBuffer)
	}

	readRequestSize -= readCounter

	streamRead := make([]byte, readRequestSize)

	if readRequestSize > 0 {
		readFromStreamCount, err := s.Stream.Read(streamRead)
		readCounter += readFromStreamCount
		if err != nil {
			return readCounter, err
		}
		s.ReadBuffer = append(s.ReadBuffer, streamRead...)
	}

	return readCounter, nil
}

func (s *RpcSession) Seek(needle []byte, bufferSize int, limit int) (index int, err error) {
	needleLength := len(needle)
	readBuffer := make([]byte, bufferSize)
	searchBuffer := make([]byte, 0, bufferSize)

	i := 0

	for i < limit {
		n, err := s.Read(readBuffer)
		if err != nil {
			return i, err
		}

		toKeep := len(searchBuffer) - needleLength
		if toKeep < len(searchBuffer) {
			toKeep = len(searchBuffer)
		}

		// increment index base when moving the searchbuffer
		i += len(searchBuffer) - toKeep

		searchBuffer = append(searchBuffer[toKeep:], readBuffer[n:]...)

		offset := findSubSlice(searchBuffer, needle)
		if offset != -1 {
			return i + offset, nil
		}
	}

	return i, fmt.Errorf("Error seeking needle, limit reached")
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
		log.Printf("Error reading header: %v", err)
	}

	debug, err := json.Marshal(header)
	fmt.Printf("\nHeader\n%v\n", debug)
}

type SessionRequestHeader struct {
	Type string                 `json:"type"`
	Args map[string]interface{} `json:"args"`
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
