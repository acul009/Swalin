package rmm

import (
	"fmt"

	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/rpc"
)

func UploadHostConfigCommandHandler[T HostConfig]() rpc.RpcCommand {
	return &uploadHostConfigCommand[T]{}
}

type uploadHostConfigCommand[T HostConfig] struct {
	Config []byte
}

func NewUploadHostCommand[T HostConfig](config *pki.SignedArtifact[T]) *uploadHostConfigCommand[T] {
	return &uploadHostConfigCommand[T]{
		Config: config.Raw(),
	}
}

func (c *uploadHostConfigCommand[T]) GetKey() string {
	var conf T
	return "upload-host-config-" + conf.GetConfigKey()
}

func (c *uploadHostConfigCommand[T]) ExecuteClient(session *rpc.RpcSession) error {
	return nil
}

func (c *uploadHostConfigCommand[T]) ExecuteServer(session *rpc.RpcSession) error {
	conf, err := pki.LoadSignedArtifact[T](c.Config, session.Verifier())
	if err != nil {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 500,
			Msg:  "Error unmarshaling config",
		})
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	//TODO
	conf.Raw()

	session.WriteResponseHeader(rpc.SessionResponseHeader{
		Code: 200,
		Msg:  "Host config saved",
	})

	return nil
}
