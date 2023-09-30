package rpc

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/quic-go/quic-go"
)

type RpcConnectionState int16

const (
	RpcConnectionOpen RpcConnectionState = iota
	RpcConnectionStopped
)

type RpcConnection struct {
	quic.Connection
	server   *RpcServer
	Uuid     uuid.UUID
	state    RpcConnectionState
	sessions map[uuid.UUID]*RpcSession
	mutex    sync.Mutex
}

func NewRpcConnection(conn quic.Connection, server *RpcServer) *RpcConnection {
	return &RpcConnection{
		Connection: conn,
		server:     server,
		state:      RpcConnectionOpen,
		Uuid:       uuid.New(),
	}
}

func (conn *RpcConnection) serve(commands *CommandCollection) error {
	conn.mutex.Lock()
	if conn.state != RpcConnectionOpen {
		conn.mutex.Unlock()
		return fmt.Errorf("RPC connection not open")
	}
	conn.mutex.Unlock()

	fmt.Println("Connection accepted, serving RPC")
	for {
		session, err := conn.AcceptSession(context.Background())

		conn.mutex.Lock()
		if conn.state != RpcConnectionOpen {
			conn.mutex.Unlock()
			return fmt.Errorf("RPC connection not open anymore")
		}
		conn.mutex.Unlock()

		if err != nil {
			log.Printf("error accepting QUIC stream: %v", err)
		}

		go session.handleIncoming(commands)
	}

}

func (conn *RpcConnection) AcceptSession(context.Context) (*RpcSession, error) {
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error accepting QUIC stream: %v", err)
	}
	return NewRpcSession(stream, conn), nil
}

func (conn *RpcConnection) OpenSession(ctx context.Context) (*RpcSession, error) {
	conn.mutex.Lock()
	if conn.state != RpcConnectionOpen {
		conn.mutex.Unlock()
		return nil, fmt.Errorf("RPC connection not open anymore")
	}
	defer conn.mutex.Unlock()

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		return nil, fmt.Errorf("error opening QUIC stream: %v", err)
	}

	return NewRpcSession(stream, conn), nil
}

func (conn *RpcConnection) Close(code quic.ApplicationErrorCode, msg string) error {
	conn.mutex.Lock()
	if conn.state != RpcConnectionOpen {
		conn.mutex.Unlock()
		return fmt.Errorf("RPC connection not open anymore")
	}
	defer conn.mutex.Unlock()

	if conn.server != nil {
		conn.server.removeConnection(conn.Uuid)
	}

	err := conn.Connection.CloseWithError(code, msg)
	return err
}
