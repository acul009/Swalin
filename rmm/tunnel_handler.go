package rmm

import (
	"fmt"
	"net"
)

type tunnelHandler struct {
}

func NewTunnelHandler() *tunnelHandler {
	return nil
}

type activeTcpTunnel struct {
	listener net.Listener
}

func (a *activeTcpTunnel) Close() error {

}

func (a *activeTcpTunnel) Run() error {
	for {
		conn, err := a.listener.Accept()
		if err != nil {
			return fmt.Errorf("error accepting connection: %w", err)
		}

	}
}

func (th *tunnelHandler) OpenTcpTunnel(tunnel *TcpTunnel) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", tunnel.ListenPort))
	if err != nil {
		return fmt.Errorf("error listening on port %d: %w", tunnel.ListenPort, err)
	}

	t := activeTcpTunnel{
		listener: listener,
	}
	return nil
}
