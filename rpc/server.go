package rpc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"rahnit-rmm/pki"
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
	rpcCommands       *CommandCollection
	state             RpcServerState
	activeConnections map[uuid.UUID]*rpcConnection
	mutex             sync.Mutex
	nonceStorage      *nonceStorage
}

type RpcServerState int16

const (
	RpcServerCreated RpcServerState = iota
	RpcServerRunning
	RpcServerStopped
)

func NewRpcServer(listenAddr string, rpcCommands *CommandCollection) (*RpcServer, error) {
	tlsConf, err := getTlsServerConfig([]TlsConnectionProto{ProtoRpc, ProtoClientLogin})
	if err != nil {
		return nil, fmt.Errorf("error getting server tls config: %w", err)
	}

	quicConf := &quic.Config{}
	listener, err := quic.ListenAddr(listenAddr, tlsConf, quicConf)
	if err != nil {
		return nil, fmt.Errorf("error creating QUIC server: %w", err)
	}

	return &RpcServer{
		listener:          listener,
		rpcCommands:       rpcCommands,
		state:             RpcServerCreated,
		activeConnections: make(map[uuid.UUID]*rpcConnection),
		mutex:             sync.Mutex{},
		nonceStorage:      NewNonceStorage(),
	}, nil
}

func (s *RpcServer) accept() (*rpcConnection, error) {
	conn, err := s.listener.Accept(context.Background())
	if err != nil {
		conn.CloseWithError(400, "")
		return nil, fmt.Errorf("error accepting QUIC connection: %w", err)
	}

	state := conn.ConnectionState()

	peerCertList := state.TLS.PeerCertificates
	if len(peerCertList) > 1 {
		conn.CloseWithError(400, "")
		return nil, fmt.Errorf("peer provided multiple certificates")
	}

	var protocol TlsConnectionProto
	var peerCert *pki.Certificate

	if len(peerCertList) == 1 {
		peerCert, err = pki.ImportCertificate(peerCertList[0])
		if err != nil {
			conn.CloseWithError(400, "")
			return nil, fmt.Errorf("error parsing peer certificate: %w", err)
		}
	} else {
		peerCert = nil
	}

	switch TlsConnectionProto(state.TLS.NegotiatedProtocol) {
	case ProtoRpc:

		if peerCert == nil {
			conn.CloseWithError(400, "")
			return nil, fmt.Errorf("peer did not provide a any certificate")
		}

		if err := pki.VerifyCertificate(peerCert); err != nil {
			conn.CloseWithError(400, "")
			return nil, fmt.Errorf("peer did not provide a valid certificate: %w", err)
		}

		protocol = ProtoRpc

	case ProtoClientLogin:

		if peerCert != nil {
			conn.CloseWithError(400, "")
			return nil, fmt.Errorf("peer provided certificate for login")
		}

		protocol = ProtoClientLogin

	default:
		conn.CloseWithError(400, "")
		return nil, fmt.Errorf("unwanted connection type: wrong tls-next protocol")

	}

	var connection *rpcConnection

	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i := 0; i < 10; i++ {
		newConnection := newRpcConnection(conn, s, RpcRoleServer, s.nonceStorage, peerCert, protocol)
		if _, ok := s.activeConnections[newConnection.Uuid]; !ok {
			connection = newConnection
			break
		}
	}

	if connection == nil {
		conn.CloseWithError(400, "")
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

		switch conn.protocol {
		case ProtoRpc:
			go conn.serve(s.rpcCommands)

		case ProtoClientLogin:
			go func() {
				err := acceptLoginRequest(conn)
				if err != nil {
					log.Printf("error accepting login request: %v", err)
				}
			}()

		default:
			log.Printf("error accepted connection has wrong protocol: %s", conn.protocol)
			continue
		}

		// TODO: check certificate

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
		go func(connection *rpcConnection) {
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
