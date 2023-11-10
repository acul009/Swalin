package rmm

import (
	"net"
	"rahnit-rmm/pki"
)

type TunnelConfig struct {
	Host pki.PublicKey
	Tcp  []tcpTunnel
}

type tcpTunnel struct {
	Name        string
	InboundPort uint16
	Target      net.TCPAddr
}

func (t *TunnelConfig) MayPublish(cert *pki.Certificate) bool {
	typ := cert.Type()
	return typ == pki.CertTypeRoot || typ == pki.CertTypeUser
}
