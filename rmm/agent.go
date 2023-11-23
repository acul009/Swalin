package rmm

import (
	"context"
	"rahnit-rmm/pki"
	"rahnit-rmm/rpc"
)

type Agent struct {
	*rpc.RpcEndpoint[*Dependencies]
}

func AgentConnect(ctx context.Context, credentials *pki.PermanentCredentials) (*Agent, error) {
	deps := &Dependencies{}

	ep, err := rpc.ConnectToUpstream(ctx, credentials, deps)
	if err != nil {
		return nil, err
	}

	a := &Agent{
		RpcEndpoint: ep,
	}

	return a, nil
}
