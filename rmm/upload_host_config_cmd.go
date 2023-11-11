package rmm

import (
	"rahnit-rmm/pki"
	"rahnit-rmm/rpc"
)

func UploadHostConfigCommandHandler[T HostConfig]() rpc.RpcCommand {
	return &uploadHostConfigCommand[T]{}
}

type uploadHostConfigCommand[T HostConfig] struct {
	Config *pki.SignedArtifact[T]
}

func NewUploadHostCommand[T HostConfig](config *pki.SignedArtifact[T]) *uploadHostConfigCommand[T] {
	return &uploadHostConfigCommand[T]{
		Config: config,
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

}
