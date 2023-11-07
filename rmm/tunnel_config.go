package rmm

import "net"

type TunnelConfig struct {
	Tcp []tcpTunnel
}

type tcpTunnel struct {
	Name        string
	InboundPort uint16
	Target      net.TCPAddr
}
