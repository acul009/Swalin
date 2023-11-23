package rpc

import (
	"context"
	"fmt"
	"rahnit-rmm/config"
	"rahnit-rmm/pki"
	"rahnit-rmm/util"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
)

type RpcEndpointState int16

const (
	RpcEndpointRunning RpcEndpointState = iota
	RpcEndpointClosed
)

type RpcEndpoint[T any] struct {
	deps  T
	conn  *RpcConnection[T]
	state RpcEndpointState
	mutex sync.Mutex
}

func ConnectToUpstream[T any](ctx context.Context, credentials pki.Credentials, deps T) (*RpcEndpoint[T], error) {
	upstreamAddr := config.V().GetString("upstream.address")
	if upstreamAddr == "" {
		return nil, fmt.Errorf("upstream address is missing")
	}

	upstreamCert, err := pki.Upstream.Get()
	if err != nil {
		return nil, fmt.Errorf("error parsing upstream certificate: %w", err)
	}

	return newRpcEndpoint[T](ctx, upstreamAddr, credentials, upstreamCert, deps)
}

func newRpcEndpoint[T any](ctx context.Context, addr string, credentials pki.Credentials, partner *pki.Certificate, deps T) (*RpcEndpoint[T], error) {
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

	rpcConn := newRpcConnection[T](quicConn, nil, RpcRoleClient, util.NewNonceStorage(), partner, ProtoRpc, credentials, nil)

	ep := &RpcEndpoint[T]{
		deps:  deps,
		conn:  rpcConn,
		state: RpcEndpointRunning,
		mutex: sync.Mutex{},
	}

	verifier, err := NewUpstreamVerify(ep)
	if err != nil {
		return nil, fmt.Errorf("error creating upstream verify: %w", err)
	}

	ep.conn.verifier = verifier

	return ep, nil
}

func (r *RpcEndpoint[T]) SendCommand(ctx context.Context, cmd RpcCommand) (util.AsyncAction, error) {
	if r == nil {
		return nil, fmt.Errorf("endpoint is nil")
	}

	err := r.ensureState(RpcEndpointRunning)
	if err != nil {
		return nil, fmt.Errorf("error mutating endpoint state: %w", err)
	}

	session, err := r.conn.OpenSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("error opening session: %w", err)
	}

	running, err := session.sendCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("error sending command: %w", err)
	}

	return running, nil
}

func (r *RpcEndpoint[T]) SendSyncCommand(ctx context.Context, cmd RpcCommand) error {
	running, err := r.SendCommand(ctx, cmd)
	if err != nil {
		return err
	}

	err = running.Wait()
	if err != nil {
		return fmt.Errorf("error executing command: %w", err)
	}

	return nil
}

func (r *RpcEndpoint[T]) SendCommandTo(ctx context.Context, to *pki.Certificate, cmd RpcCommand) (util.AsyncAction, error) {
	if r == nil {
		return nil, fmt.Errorf("endpoint is nil")
	}

	encrypt, err := newE2eEncryptCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("error preparing encryption: %w", err)
	}

	if r.conn.verifier == nil {
		return nil, fmt.Errorf("verifier is nil")
	}

	_, err = r.conn.verifier.Verify(to)
	if err != nil {
		return nil, fmt.Errorf("error verifying target certificate: %w", err)
	}

	forward := newForwardCommand(to, encrypt)
	return r.SendCommand(ctx, forward)
}

func (r *RpcEndpoint[T]) SendSyncCommandTo(ctx context.Context, to *pki.Certificate, cmd RpcCommand) error {
	running, err := r.SendCommandTo(ctx, to, cmd)
	if err != nil {
		return err
	}

	err = running.Wait()
	if err != nil {
		return fmt.Errorf("error executing command: %w", err)
	}

	return nil
}

func (r *RpcEndpoint[T]) Close(code quic.ApplicationErrorCode, msg string) error {
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

func (r *RpcEndpoint[T]) ServeRpc(commands *CommandCollection[T]) error {
	if r == nil {
		return fmt.Errorf("endpoint is nil")
	}

	return r.conn.serveRpc(commands)
}

func (r *RpcEndpoint[T]) Credentials() *pki.PermanentCredentials {
	credentials := r.conn.credentials

	if credentials == nil {
		panic("credentials is nil")
	}

	perm, ok := credentials.(*pki.PermanentCredentials)
	if !ok {
		panic("credentials is not permanent")
	}

	return perm
}

func (r *RpcEndpoint[T]) ensureState(state RpcEndpointState) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.state != state {
		return fmt.Errorf("RPC endpoint not in state %v", state)
	}
	return nil
}

func (r *RpcEndpoint[T]) mutateState(from RpcEndpointState, to RpcEndpointState) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.state != from {
		return fmt.Errorf("RPC endpoint not in state %v", from)
	}
	r.state = to
	return nil
}
