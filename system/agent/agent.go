package agent

import (
	"errors"
	"fmt"
	"log"

	"github.com/rahn-it/svalin/config"
	"github.com/rahn-it/svalin/rpc"
)

type Agent struct {
}

func Connect(profile *config.Profile) (*Agent, error) {
	return nil, errors.New("not implemented")
}

func (a *Agent) Run() error {
	return errors.New("not implemented")
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

	err = initAgentConfig(scope, addr, initInfo)
	if err != nil {
		return fmt.Errorf("error initializing agent config: %w", err)
	}

	return errors.New("not implemented")
}
