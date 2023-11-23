package rmm

import (
	"context"
	"fmt"
	"log"
	"net"
	"rahnit-rmm/pki"
	"rahnit-rmm/util"
	"sync"
)

type tunnelHandler struct {
	cli *Client

	TcpTunnels util.ObservableMap[*tcpTunnelConnectionDetails, *ActiveTcpTunnel]
}

func newTunnelHandler(cli *Client) *tunnelHandler {
	return &tunnelHandler{
		cli:        cli,
		TcpTunnels: util.NewObservableMap[*tcpTunnelConnectionDetails, *ActiveTcpTunnel](),
	}
}

func (th *tunnelHandler) OpenTcpTunnel(device *pki.Certificate, tunnel *TcpTunnel) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", tunnel.ListenPort))
	if err != nil {
		return fmt.Errorf("error listening on port %d: %w", tunnel.ListenPort, err)
	}

	t := &ActiveTcpTunnel{
		tcpTunnelConnectionDetails: tcpTunnelConnectionDetails{*tunnel, device},
		cli:                        th.cli,
		listener:                   listener,
	}

	err = t.Run()
	if err != nil {
		return fmt.Errorf("error running tunnel: %w", err)
	}

	t.onClose = func() {
		th.TcpTunnels.Delete(&t.tcpTunnelConnectionDetails)
	}

	th.TcpTunnels.Set(&t.tcpTunnelConnectionDetails, t)

	return nil
}

type tcpTunnelConnectionDetails struct {
	TcpTunnel
	Device *pki.Certificate
}

type ActiveTcpTunnel struct {
	tcpTunnelConnectionDetails
	cli             *Client
	device          *pki.Certificate
	listener        net.Listener
	openConnections map[net.Addr]util.AsyncAction
	mutex           sync.Mutex
	onClose         func()
}

func (a *ActiveTcpTunnel) Close() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.onClose != nil {
		a.onClose()
		a.onClose = nil
	}

	for _, c := range a.openConnections {
		c.Close()
	}

	return a.listener.Close()
}

func (a *ActiveTcpTunnel) Run() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", a.ListenPort))
	if err != nil {
		return fmt.Errorf("error listening on port %d: %w", a.ListenPort, err)
	}

	a.listener = listener

	go func() {
		err := a.acceptAndForward()
		if err != nil {
			log.Printf("Error accepting and forwarding: %v", err)
		}
	}()

	return nil
}

func (a *ActiveTcpTunnel) acceptAndForward() error {
	conn, err := a.listener.Accept()
	if err != nil {
		return fmt.Errorf("error accepting connection: %w", err)
	}

	conn.RemoteAddr()

	cmd := NewTcpForwardCommand(a.Target, conn)

	running, err := a.cli.SendCommandTo(context.Background(), a.device, cmd)
	if err != nil {
		return fmt.Errorf("error running command: %w", err)
	}

	a.mutex.Lock()
	a.openConnections[conn.RemoteAddr()] = running
	a.mutex.Unlock()

	go func() {
		a.mutex.Lock()
		defer a.mutex.Unlock()

		delete(a.openConnections, conn.RemoteAddr())

		err := running.Wait()
		if err != nil {
			log.Printf("Error running command: %v", err)
		}

		err = conn.Close()
		if err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}()

	return nil
}
