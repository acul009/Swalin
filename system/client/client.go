package client

import (
	"errors"
	"fmt"
	"log"

	"github.com/rahn-it/svalin/config"
	"github.com/rahn-it/svalin/pki"
)

type Client struct {
}

func OpenClient(profile *config.Profile) (*Client, error) {
	log.Printf("opening client for profile %s", profile.Name())
	return nil, errors.New("not implemented")
}

func SetupClient(
	profile *config.Profile,
	root *pki.Certificate,
	upstream *pki.Certificate,
	credentials *pki.PermanentCredentials,
	password []byte,
	serverAddr string,
) error {
	err := initClientConfig(profile.Scope().Scope("client"), root, upstream, credentials, password, serverAddr)
	if err != nil {
		return fmt.Errorf("failed to initialize client config: %w", err)
	}

	return nil
}
