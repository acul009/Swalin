package rpc

import (
	"context"
	"fmt"
	"rahnit-rmm/config"
	"rahnit-rmm/pki"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
)

type RpcEndpointState int16

const (
	RpcEndpointRunning RpcEndpointState = iota
	RpcEndpointClosed
)

type RpcEndpoint struct {
	conn  *RpcConnection
	state RpcEndpointState
	mutex sync.Mutex
}

func ConnectToUpstream(ctx context.Context, credentials pki.Credentials) (*RpcEndpoint, error) {
	upstreamAddr := config.V().GetString("upstream.address")
	if upstreamAddr == "" {
		return nil, fmt.Errorf("upstream address is missing")
	}

	upstreamCert, err := pki.Upstream.Get()
	if err != nil {
		return nil, fmt.Errorf("error parsing upstream certificate: %w", err)
	}

	return newRpcEndpoint(ctx, upstreamAddr, credentials, upstreamCert)
}

func newRpcEndpoint(ctx context.Context, addr string, credentials pki.Credentials, partner *pki.Certificate) (*RpcEndpoint, error) {
	if addr == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	if partner == nil {
		return nil, fmt.Errorf("partner cannot be nil")
	}

	tlsConf := getTlsClientConfig(ProtoRpc, credentials)

	quicConf := &quic.Config{
		KeepAlivePeriod: 30 * time.Second,
	}

	quicConn, err := quic.DialAddr(context.Background(), addr, tlsConf, quicConf)
	if err != nil {
		qErr, ok := err.(*quic.TransportError)
		if ok && uint8(qErr.ErrorCode) == 120 {
			return nil, fmt.Errorf("server not in rpc mode (not initialized yet?): %w", err)
		}
		return nil, fmt.Errorf("error creating QUIC connection: %w", err)
	}

	rpcConn := newRpcConnection(quicConn, nil, RpcRoleClient, NewNonceStorage(), partner, ProtoRpc, credentials, nil)

	verifier, err := NewUpstreamVerify()
	if err != nil {
		return nil, fmt.Errorf("error creating upstream verify: %w", err)
	}

	rpcConn.verifier = verifier

	return &RpcEndpoint{
		conn:  rpcConn,
		state: RpcEndpointRunning,
		mutex: sync.Mutex{},
	}, nil
}

func (r *RpcEndpoint) SendCommand(ctx context.Context, cmd RpcCommand) error {
	err := r.ensureState(RpcEndpointRunning)
	if err != nil {
		return fmt.Errorf("error mutating endpoint state: %w", err)
	}

	session, err := r.conn.OpenSession(ctx)
	if err != nil {
		return fmt.Errorf("error opening session: %w", err)
	}

	return session.SendCommand(cmd)
}

func (r *RpcEndpoint) SendCommandTo(ctx context.Context, to *pki.Certificate, cmd RpcCommand) error {
	encrypt, err := newE2eEncryptCommand(cmd)
	if err != nil {
		return fmt.Errorf("error preparing encryption: %w", err)
	}
	forward := newForwardCommand(to, encrypt)
	return r.SendCommand(ctx, forward)
}

func (r *RpcEndpoint) Close(code quic.ApplicationErrorCode, msg string) error {
	err := r.mutateState(RpcEndpointRunning, RpcEndpointClosed)
	if err != nil {
		return fmt.Errorf("error mutating endpoint state: %w", err)
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	err = r.conn.Close(code, msg)
	if err != nil {
		return fmt.Errorf("error closing connection: %w", err)
	}

	return nil
}

func (r *RpcEndpoint) ServeRpc(commands *CommandCollection) error {
	return r.conn.serveRpc(commands)
}

func (r *RpcEndpoint) ensureState(state RpcEndpointState) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.state != state {
		return fmt.Errorf("RPC endpoint not in state %v", state)
	}
	return nil
}

func (r *RpcEndpoint) mutateState(from RpcEndpointState, to RpcEndpointState) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.state != from {
		return fmt.Errorf("RPC endpoint not in state %v", from)
	}
	r.state = to
	return nil
}
