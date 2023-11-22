package rmm

import "rahnit-rmm/pki"

type tunnelHandler struct {
}

func NewTunnelHandler() *tunnelHandler {
	return nil
}

type activeTunnel struct {
}

func (t *tunnelHandler) OpenTunnel(tunnel *pki.SignedArtifact[*TunnelConfig]) error {
	return nil
}
