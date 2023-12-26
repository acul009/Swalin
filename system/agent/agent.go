package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rahn-it/svalin/config"
	"github.com/rahn-it/svalin/rmm"
	"github.com/rahn-it/svalin/rpc"
	"github.com/rahn-it/svalin/system"
)

type Agent struct {
	ep           *rpc.RpcEndpoint
	profile      *config.Profile
	agent_config *agentConfig
	commands     *rpc.CommandCollection
}

func Connect(profile *config.Profile) (*Agent, error) {
	scope := profile.Scope()

	config, err := openClientConfig(scope.Scope("agent"))
	if err != nil {
		return nil, fmt.Errorf("error opening client config: %w", err)
	}

	verifier := system.NewUpstreamVerifier(config.Upstream(), config.Root(), nil)

	ep, err := rpc.ConnectToServer(context.Background(), config.ServerAddr(), config.Credentials(), config.Upstream(), verifier)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	verifier.SetEndPoint(ep)

	commands := rpc.NewCommandCollection(
		rmm.MonitorSystemCommandHandler,
		rmm.MonitorProcessesCommandHandler,
		rmm.MonitorServicesCommandHandler,
		rmm.KillProcessCommandHandler,
		rmm.RemoteShellCommandHandler,
	)

	a := &Agent{
		ep:           ep,
		profile:      profile,
		agent_config: config,
		commands:     commands,
	}

	return a, nil
}

func (a *Agent) Run() error {
	return a.ep.ServeRpc(rpc.NewCommandCollection(
		rpc.CreateE2eDecryptCommandHandler(a.commands),
	))
}

func Init(profile *config.Profile) error {
	scope := profile.Scope()

	found, err := checkForAgentConfig(scope.Scope("agent"))
	if err != nil {
		return fmt.Errorf("error checking for agent config: %w", err)
	}
	if found {
		return nil
	}

	addr := profile.Config().String("agent.address")
	if addr == "" {
		return fmt.Errorf("agent address not set")
	}
	log.Printf("Starting enrollment with server at %s", addr)

	initInfo, err := rpc.EnrollWithUpstream(addr)
	if err != nil {
		return fmt.Errorf("error enrolling with server: %w", err)
	}

	log.Printf("Received certificate from server")

	err = initAgentConfig(scope.Scope("agent"), addr, initInfo)
	if err != nil {
		return fmt.Errorf("error initializing agent config: %w", err)
	}

	time.Sleep(5 * time.Second)

	return nil
}
