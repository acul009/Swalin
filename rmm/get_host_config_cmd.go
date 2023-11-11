package rmm

import (
	"fmt"
	"rahnit-rmm/pki"
	"rahnit-rmm/rpc"
)

func GetHostConfigCommandHandler[T HostConfig]() rpc.RpcCommand {
	return &GetConfigCommand[T]{}
}

type GetConfigCommand[T HostConfig] struct {
	Host   *pki.PublicKey
	config *pki.SignedArtifact[T]
}

func NewGetConfigCommand[T HostConfig](host *pki.PublicKey) *GetConfigCommand[T] {
	return &GetConfigCommand[T]{
		Host: host,
	}
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

	err = rpc.WriteMessage[[]byte](session, artifact.Raw())
	if err != nil {
		return fmt.Errorf("error writing artifact: %w", err)
	}

	return nil
}

func (c *GetConfigCommand[T]) ExecuteClient(session *rpc.RpcSession) error {
	raw := make([]byte, 0)
	err := rpc.ReadMessage[[]byte](session, raw)
	if err != nil {
		return fmt.Errorf("error receiving tunnel config: %w", err)
	}

	conf, err := pki.LoadSignedArtifact[T](raw, session.Verifier())
	if err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	*c.config = *conf

	return nil
}

func (c *GetConfigCommand[T]) Config() T {
	return c.config.Artifact()
}
