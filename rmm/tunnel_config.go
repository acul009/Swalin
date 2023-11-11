package rmm

import (
	"net"
	"rahnit-rmm/pki"
)

var _ HostConfig = (*TunnelConfig)(nil)

type TunnelConfig struct {
	Host *pki.PublicKey
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

func (t *TunnelConfig) GetHost() *pki.PublicKey {
	return t.Host
}

func (t *TunnelConfig) GetConfigKey() string {
	return "tunnel-config"
}

func (t *TunnelConfig) MayAccess(cert *pki.Certificate) bool {
	typ := cert.Type()
	if typ == pki.CertTypeRoot || typ == pki.CertTypeUser {
		return true
	}

	return cert.GetPublicKey().Equal(t.Host)
}
