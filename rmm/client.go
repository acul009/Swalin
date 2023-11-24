package rmm

import (
	"context"
	"rahnit-rmm/pki"
	"rahnit-rmm/rpc"
)

type Client struct {
	*rpc.RpcEndpoint
	tunnelHandler *tunnelHandler
}

func ClientConnect(ctx context.Context, credentials *pki.PermanentCredentials) (*Client, error) {
	ep, err := rpc.ConnectToUpstream(ctx, credentials)
	if err != nil {
		return nil, err
	}

	c := &Client{
		RpcEndpoint: ep,
	}

	th := newTunnelHandler(c)
	c.tunnelHandler = th

	return c, nil
}

func (c *Client) Tunnels() *tunnelHandler {
	return c.tunnelHandler
}
