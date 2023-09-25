package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/quic-go/quic-go"
)

type RpcServer struct {
	listener        *quic.Listener
	commands        *CommandCollection
	registeredNodes map[string]*NodeConnection
}

func NewRpcServer(listener *quic.Listener, commands *CommandCollection) *RpcServer {
	return &RpcServer{
		listener:        listener,
		commands:        commands,
		registeredNodes: make(map[string]*NodeConnection),
	}
}

type NodeConnection struct {
	quic.Connection
	server   *RpcServer
	nodeName string
}

func NewNodeConnection(conn quic.Connection, server *RpcServer) *NodeConnection {
	return &NodeConnection{
		Connection: conn,
		server:     server,
		nodeName:   "",
	}
}

func (n *NodeConnection) Close() {
	delete(n.server.registeredNodes, n.nodeName)
	n.Close()
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

func (s *RpcServer) Accept() (*NodeConnection, error) {
	conn, err := s.listener.Accept(context.Background())
	if err != nil {
		return nil, err
	}
	return NewNodeConnection(conn, s), nil
}

func (s *RpcServer) Run() error {
	fmt.Println("Starting RPC server")
	for {
		conn, err := s.Accept()
		if err != nil {
			log.Printf("Error accepting QUIC connection: %v", err)
			continue
		}

		go ServeConnection(conn, s.commands)
	}
}

func ServeConnection(conn *NodeConnection, commands *CommandCollection) {
	fmt.Println("Connection accepted, serving RPC")
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Printf("Error accepting QUIC stream: %v", err)
			return
		}
		go handleSession(NewRpcSession(stream, conn), commands)
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
		newErr := fmt.Errorf("Error handling request: %v", err)
		log.Printf("%v\n", newErr)
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
