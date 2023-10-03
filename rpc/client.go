package rpc

import (
	"context"
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
		return nil, fmt.Errorf("error creating QUIC client: %v", err)
	}
	rpcConn := NewRpcConnection(conn, nil, RpcRoleClient)
	return &RpcClient{
		conn:  rpcConn,
		state: RpcClientRunning,
		mutex: sync.Mutex{},
	}, nil
}

func (c *RpcClient) SendCommand(ctx context.Context, cmd RpcCommand) error {
	c.mutex.Lock()
	if c.state != RpcClientRunning {
		c.mutex.Unlock()
		return fmt.Errorf("RPC client not running anymore")
	}
	c.mutex.Unlock()
	session, err := c.conn.OpenSession(ctx)
	if err != nil {
		return err
	}

	err = session.SendCommand(cmd)
	if err != nil {
		return fmt.Errorf("error sending command: %v", err)
	}

	err = session.Close()
	if err != nil {
		return fmt.Errorf("error closing session: %v", err)
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
		return fmt.Errorf("error closing connection: %v", err)
	}
	return nil
}
