package rmm

import (
	"fmt"

	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/rpc"
	"github.com/rahn-it/svalin/util"
)

func CreateHostConfigCommandHandler[T HostConfig](source util.ObservableMap[string, *pki.SignedArtifact[T]]) func() rpc.RpcCommand {
	return func() rpc.RpcCommand {
		return &GetConfigCommand[T]{
			sourceMap: source,
		}
	}
}

type GetConfigCommand[T HostConfig] struct {
	Host      *pki.Certificate
	config    util.UpdateableObservable[T]
	sourceMap util.ObservableMap[string, *pki.SignedArtifact[T]]
}

func NewGetConfigCommand[T HostConfig](host *pki.Certificate, config util.UpdateableObservable[T]) *GetConfigCommand[T] {
	return &GetConfigCommand[T]{
		Host:   host,
		config: config,
	}
}

func (c *GetConfigCommand[T]) GetKey() string {
	var conf T
	return "get-host-config-" + conf.GetConfigKey()
}

func (c *GetConfigCommand[T]) ExecuteServer(session *rpc.RpcSession) error {
	// TODO
	return fmt.Errorf("not implemented")
}

func (c *GetConfigCommand[T]) ExecuteClient(session *rpc.RpcSession) error {
	raw := make([]byte, 0)

	for {

		err := rpc.ReadMessage[[]byte](session, raw)
		if err != nil {
			return fmt.Errorf("error receiving tunnel config: %w", err)
		}

		conf, err := pki.LoadSignedArtifact[T](raw, session.Verifier())
		if err != nil {
			return fmt.Errorf("error unmarshaling config: %w", err)
		}

		c.config.Update(func(t T) T {
			return conf.Artifact()
		})

	}
}
