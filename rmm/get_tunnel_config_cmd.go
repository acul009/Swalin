package rmm

import (
	"context"
	"fmt"
	"rahnit-rmm/config"
	"rahnit-rmm/ent/device"
	"rahnit-rmm/ent/tunnelconfig"
	"rahnit-rmm/pki"
	"rahnit-rmm/rpc"
)

type GetTunnelConfigCommand struct {
	Host   pki.PublicKey
	config *pki.SignedArtifact[*TunnelConfig]
}

func (c *GetTunnelConfigCommand) GetKey() string {
	return "get_tunnel_config"
}

func (c *GetTunnelConfigCommand) ExecuteServer(session *rpc.RpcSession) error {
	db := config.DB()

	savedConfig, err := db.TunnelConfig.Query().Where(tunnelconfig.HasDeviceWith(device.PublicKey(c.Host.Base64Encode()))).Only(context.Background())
	if err != nil {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 404,
			Msg:  "Tunnel config not found",
		})
		return fmt.Errorf("error querying tunnel config: %w", err)
	}

	config := &TunnelConfig{}
	artifact, err := pki.LoadSignedArtifact[*TunnelConfig](savedConfig.Config, session.Verifier(), config)
	if err != nil {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 500,
			Msg:  "Error unmarshaling config",
		})
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	session.WriteResponseHeader(rpc.SessionResponseHeader{
		Code: 200,
		Msg:  "Tunnel config found",
	})

	err = rpc.WriteMessage[*pki.SignedArtifact[*TunnelConfig]](session, artifact)
	if err != nil {
		return fmt.Errorf("error writing artifact: %w", err)
	}

	return nil
}

func (c *GetTunnelConfigCommand) ExecuteClient(session *rpc.RpcSession) error {
	err := rpc.ReadMessage[*pki.SignedArtifact[*TunnelConfig]](session, c.config)
	if err != nil {
		return fmt.Errorf("error receiving tunnel config: %w", err)
	}

	return nil
}

func (c *GetTunnelConfigCommand) Config() *TunnelConfig {
	return c.config.Artifact()
}
