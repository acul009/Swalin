package rmm

import (
	"fmt"
	"rahnit-rmm/pki"
	"rahnit-rmm/rpc"
)

type GetConfigCommand[T HostConfig] struct {
	Host   *pki.PublicKey
	config *pki.SignedArtifact[T]
}

func (c *GetConfigCommand[T]) GetKey() string {
	var conf T
	return "get-host-config-" + conf.GetConfigKey()
}

func (c *GetConfigCommand[T]) ExecuteServer(session *rpc.RpcSession) error {

	artifact, err := LoadHostConfigFromDB[T](c.Host, session.Verifier())
	if err != nil {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 500,
			Msg:  "Error unmarshaling config",
		})
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	if !artifact.Artifact().MayAccess(session.Partner()) {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 403,
			Msg:  "Not authorized",
		})
		return fmt.Errorf("not authorized")
	}

	session.WriteResponseHeader(rpc.SessionResponseHeader{
		Code: 200,
		Msg:  "Tunnel config found",
	})

	err = rpc.WriteMessage[*pki.SignedArtifact[T]](session, artifact)
	if err != nil {
		return fmt.Errorf("error writing artifact: %w", err)
	}

	return nil
}

func (c *GetConfigCommand[T]) ExecuteClient(session *rpc.RpcSession) error {
	err := rpc.ReadMessage[*pki.SignedArtifact[T]](session, c.config)
	if err != nil {
		return fmt.Errorf("error receiving tunnel config: %w", err)
	}

	return nil
}

func (c *GetConfigCommand[T]) Config() T {
	return c.config.Artifact()
}
