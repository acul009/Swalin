package rmm

import (
	"context"
	"errors"
	"fmt"

	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/rpc"
)

type Agent struct {
	*rpc.RpcEndpoint
}

func AgentConnect(ctx context.Context, credentials *pki.PermanentCredentials) (*Agent, error) {

	// ep, err := rpc.ConnectToUpstream(ctx, credentials)
	// if err != nil {
	// 	return nil, fmt.Errorf("unable to connect to upstream server: %w", err)
	// }

	// a := &Agent{
	// 	RpcEndpoint: ep,
	// }

	// return a, nil

	return nil, errors.New("deprecated")
}

func (a *Agent) Run() error {

	cmdCollection := rpc.NewCommandCollection(
		rpc.CreateE2eDecryptCommandHandler(rpc.NewCommandCollection(
			rpc.PingHandler,
			MonitorSystemCommandHandler,
			MonitorProcessesCommandHandler,
			MonitorServicesCommandHandler,
			RemoteShellCommandHandler,
			KillProcessCommandHandler,
		)),
	)

	err := a.ServeRpc(cmdCollection)
	if err != nil {
		return fmt.Errorf("error serving rpc: %w", err)
	}

	return nil
}

func (a *Agent) Close() error {
	return a.RpcEndpoint.Close(200, "Shutdown")
}
