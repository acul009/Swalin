package agent

import (
	"errors"

	"github.com/rahn-it/svalin/config"
)

type Agent struct {
}

func Connect(profile *config.Profile) (*Agent, error) {
	return nil, errors.New("not implemented")
}

func (a *Agent) Run() error {
	return errors.New("not implemented")
}
