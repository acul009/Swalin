package rmm

import (
	"fmt"
	"log"
	"rahnit-rmm/pki"
	"rahnit-rmm/rpc"
	"rahnit-rmm/util"
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
	requested := c.Host.GetPublicKey().Base64Encode()

	artifact, ok := c.sourceMap.Get(requested)
	if !ok {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 404,
			Msg:  "Tunnel config not found",
		})
		return nil
	}

	if !artifact.Artifact().MayAccess(session.Partner()) {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 403,
			Msg:  "Not authorized",
		})
		log.Printf("%s tried to get forbidden config", session.Partner())
		return fmt.Errorf("not authorized")
	}

	session.WriteResponseHeader(rpc.SessionResponseHeader{
		Code: 200,
		Msg:  "Tunnel config found",
	})

	err := rpc.WriteMessage[[]byte](session, artifact.Raw())
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}

	errChan := make(chan error)

	unsubscribe := c.sourceMap.Subscribe(
		func(key string, artifact *pki.SignedArtifact[T]) {
			if key != requested {
				return
			}
			err := rpc.WriteMessage[[]byte](session, artifact.Raw())
			if err != nil {
				errChan <- fmt.Errorf("error writing message: %w", err)
			}
		},
		func(_ string, _ *pki.SignedArtifact[T]) {
			// TODO ???
			return
		},
	)

	defer unsubscribe()

	return <-errChan
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
