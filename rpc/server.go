package rpc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"rahnit-rmm/connection"
	"sync"

	"github.com/google/uuid"
	"github.com/quic-go/quic-go"
)

type rpcNotRunningError struct {
}

func (e rpcNotRunningError) Error() string {
	return fmt.Errorf("rpc not running anymore").Error()
}

var ErrRpcNotRunning = rpcNotRunningError{}

func (e rpcNotRunningError) Is(target error) bool {
	_, ok := target.(rpcNotRunningError)
	return ok
}

type RpcServer struct {
	listener          *quic.Listener
	unsecureCommands  *CommandCollection
	tlsCommands       *CommandCollection
	state             RpcServerState
	activeConnections map[uuid.UUID]*RpcConnection
	mutex             sync.Mutex
	nonceStorage      *nonceStorage
}

type RpcServerState int16

const (
	RpcServerCreated RpcServerState = iota
	RpcServerRunning
	RpcServerStopped
)

func NewRpcServer(addr string, unsecureCommands *CommandCollection, tlsCommands *CommandCollection) (*RpcServer, error) {
	listener, err := connection.CreateServer(addr)
	if err != nil {
		return nil, fmt.Errorf("error creating QUIC server: %w", err)
	}

	return &RpcServer{
		listener:          listener,
		unsecureCommands:  unsecureCommands,
		tlsCommands:       tlsCommands,
		state:             RpcServerCreated,
		activeConnections: make(map[uuid.UUID]*RpcConnection),
		mutex:             sync.Mutex{},
		nonceStorage:      NewNonceStorage(),
	}, nil
}

func (s *RpcServer) accept() (*RpcConnection, error) {
	conn, err := s.listener.Accept(context.Background())
	if err != nil {
		return nil, err
	}
	var connection *RpcConnection

	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i := 0; i < 10; i++ {
		newConnection := NewRpcConnection(conn, s, RpcRoleServer, s.nonceStorage)
		if _, ok := s.activeConnections[newConnection.Uuid]; !ok {
			connection = newConnection
			break
		}
	}
	if connection == nil {
		return nil, fmt.Errorf("multiple uuid collisions, this should mathematically be impossible")
	}
	s.activeConnections[connection.Uuid] = connection
	return connection, nil
}

func (s *RpcServer) removeConnection(uuid uuid.UUID) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.activeConnections, uuid)
}

func (s *RpcServer) Run() error {
	fmt.Println("Starting RPC server")
	s.mutex.Lock()
	if s.state != RpcServerCreated {
		s.mutex.Unlock()
		return fmt.Errorf("RPC server already running")
	}
	s.state = RpcServerRunning
	s.mutex.Unlock()
	for {
		conn, err := s.accept()

		s.mutex.Lock()
		if s.state != RpcServerRunning {
			s.mutex.Unlock()
			return rpcNotRunningError{}
		}
		s.mutex.Unlock()

		if err != nil {
			log.Printf("error accepting QUIC connection: %v", err)
			continue
		}

		chains := conn.ConnectionState().TLS.VerifiedChains
		if len(chains) > 0 {
			log.Printf("new connection using mTLS")
			go conn.serve(s.tlsCommands)
		} else {
			certs := conn.ConnectionState().TLS.PeerCertificates
			if len(certs) > 0 {
				log.Printf("Could not verify client certificate")
			} else {
				log.Printf("new unverified connection")
				go conn.serve(s.unsecureCommands)
			}
		}

	}
}

func (s *RpcServer) Close(code quic.ApplicationErrorCode, msg string) error {

	// lock server before closing
	s.mutex.Lock()
	if s.state != RpcServerRunning {
		s.mutex.Unlock()
		return fmt.Errorf("RPC server not running")
	}
	s.state = RpcServerStopped
	connectionsToClose := s.activeConnections
	s.mutex.Unlock()

	// tell all connections to close
	errChan := make(chan error)
	wg := sync.WaitGroup{}

	errorList := make([]error, 0)

	for _, connection := range connectionsToClose {
		wg.Add(1)
		go func(connection *RpcConnection) {
			err := connection.Close(code, msg)
			if err != nil {
				errChan <- err
			}
			wg.Done()
		}(connection)
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
		err = fmt.Errorf("error closing connections: %w", errors.Join(errorList...))
	}

	s.listener.Close()

	return err
}
