package rpc

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"rahnit-rmm/connection"
	"sync"

	"github.com/quic-go/quic-go"
)

type RpcClientState int16

const (
	RpcClientRunning RpcClientState = iota
	RpcClientClosed
)

type RpcClient struct {
	conn  *RpcConnection
	state RpcClientState
	mutex sync.Mutex
}

func NewRpcClient(ctx context.Context, addr string) (*RpcClient, error) {
	conn, err := connection.CreateClient(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("error creating QUIC client: %w", err)
	}
	rpcConn := NewRpcConnection(conn, nil, RpcRoleClient, NewNonceStorage())
	return &RpcClient{
		conn:  rpcConn,
		state: RpcClientRunning,
		mutex: sync.Mutex{},
	}, nil
}

func (c *RpcClient) SendCommand(ctx context.Context, receiver *ecdsa.PublicKey, cmd RpcCommand) error {
	c.mutex.Lock()
	if c.state != RpcClientRunning {
		c.mutex.Unlock()
		return fmt.Errorf("RPC client not running anymore")
	}
	c.mutex.Unlock()
	session, err := c.conn.OpenSession(ctx)
	if err != nil {
		return fmt.Errorf("error opening session: %w", err)
	}

	err = session.SendCommand(receiver, cmd)
	if err != nil {
		return fmt.Errorf("error sending command: %w", err)
	}

	err = session.Close()
	if err != nil {
		return fmt.Errorf("error closing session: %w", err)
	}

	return nil
}

func (c *RpcClient) Close(code quic.ApplicationErrorCode, msg string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.state != RpcClientRunning {
		return fmt.Errorf("RPC client not running anymore")
	}

	c.state = RpcClientClosed

	err := c.conn.Close(code, msg)
	if err != nil {
		return fmt.Errorf("error closing connection: %w", err)
	}
	return nil
}
