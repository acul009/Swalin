package rpc

import (
	"context"
	"fmt"
	"rahnit-rmm/connection"
)

type RpcClientState int16

const (
	RpcClientCreated RpcClientState = iota
)

type RpcClient struct {
	conn  *RpcConnection
	state RpcClientState
}

func NewRpcClient(ctx context.Context, addr string) (*RpcClient, error) {
	conn, err := connection.CreateClient(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("error creating QUIC client: %v", err)
	}
	rpcConn := NewRpcConnection(conn, nil)
	return &RpcClient{
		conn:  rpcConn,
		state: RpcClientCreated,
	}, nil
}

func (c *RpcClient) SendCommand(ctx context.Context, cmd RpcCommand) error {
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

func (c *RpcClient) Close() error {

}
