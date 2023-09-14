package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/quic-go/quic-go"
)

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

func ServeSession(conn quic.Connection, commands *CommandCollection) {
	fmt.Println("Connection accepted, serving RPC")
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Printf("Error accepting QUIC stream: %v", err)
			return
		}
		go handleSession(NewRpcSession(stream), commands)
	}

}

func handleSession(session *RpcSession, commands *CommandCollection) {
	header, err := ReadRequestHeader(session)
	if err != nil {
		log.Printf("Error reading header: %v\n", err)
	}

	debug, err := json.Marshal(header)
	fmt.Printf("\nHeader\n%v\n", string(debug))
	err = commands.handleRequest(header, session)
	if err != nil {
		log.Printf("Error handling request: %v\n", err)
	}
}

var headerStop = []byte("\n")
var headerDelimiter = "|"

func ReadRequestHeader(session *RpcSession) (SessionRequestHeader, error) {
	headerData, err := session.ReadUntil(headerStop, 1024, 65536)
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

func ReadResponseHeader(session *RpcSession) (SessionResponseHeader, error) {
	headerData, err := session.ReadUntil(headerStop, 1024, 65536)
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

	return decodedHeader, nil
}
