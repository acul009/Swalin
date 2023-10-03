package rpc

import (
	"context"
	"errors"
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
	server         *RpcServer
	Uuid           uuid.UUID
	state          RpcConnectionState
	activeSessions map[uuid.UUID]*RpcSession
	mutex          sync.Mutex
}

func NewRpcConnection(conn quic.Connection, server *RpcServer) *RpcConnection {
	return &RpcConnection{
		Connection:     conn,
		server:         server,
		state:          RpcConnectionOpen,
		activeSessions: make(map[uuid.UUID]*RpcSession),
		Uuid:           uuid.New(),
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
	var session *RpcSession = nil

	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	for i := 0; i < 10; i++ {
		newSession := NewRpcSession(stream, conn)
		if _, ok := conn.activeSessions[newSession.Uuid]; !ok {
			session = newSession
			break
		}
	}

	if session == nil {
		return nil, fmt.Errorf("Multiple UUID collisions, this should mathematically be impossible")
	}

	conn.activeSessions[session.Uuid] = session

	return session, nil
}

func (conn *RpcConnection) OpenSession(ctx context.Context) (*RpcSession, error) {
	conn.mutex.Lock()
	if conn.state != RpcConnectionOpen {
		conn.mutex.Unlock()
		return nil, fmt.Errorf("RPC connection not open anymore")
	}
	conn.mutex.Unlock()

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		return nil, fmt.Errorf("error opening QUIC stream: %v", err)
	}

	return NewRpcSession(stream, conn), nil
}

func (conn *RpcConnection) removeSession(uuid uuid.UUID) {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	delete(conn.activeSessions, uuid)
}

func (conn *RpcConnection) Close(code quic.ApplicationErrorCode, msg string) error {
	conn.mutex.Lock()
	if conn.state != RpcConnectionOpen {
		conn.mutex.Unlock()
		return fmt.Errorf("RPC connection not open anymore")
	}
	conn.state = RpcConnectionStopped
	sessionsToClose := conn.activeSessions
	conn.mutex.Unlock()

	// tell all connections to close
	errChan := make(chan error)
	wg := sync.WaitGroup{}

	errorList := make([]error, 0)

	for _, session := range sessionsToClose {
		wg.Add(1)
		go func(session *RpcSession) {
			err := session.Close()
			if err != nil {
				errChan <- err
			}
			wg.Done()
		}(session)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		errorList = append(errorList, err)
	}

	var err error = nil
	if len(errorList) > 0 {
		err = fmt.Errorf("error closing sessions: %w", errors.Join(errorList...))
	}

	if conn.server != nil {
		conn.server.removeConnection(conn.Uuid)
	}

	err = conn.Connection.CloseWithError(code, msg)
	return err
}
