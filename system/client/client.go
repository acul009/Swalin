package client

import (
	"context"
	"fmt"

	"github.com/rahn-it/svalin/config"
	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/rmm"
	"github.com/rahn-it/svalin/rpc"
	"github.com/rahn-it/svalin/system"
	"github.com/rahn-it/svalin/util"
)

type Client struct {
	profile      *config.Profile
	clientConfig *clientConfig
	ep           *rpc.RpcEndpoint
}

func OpenClient(profile *config.Profile, password []byte) (*Client, error) {
	scope := profile.Scope()
	clientConfig, err := openClientConfig(scope.Scope("client"), password)
	if err != nil {
		return nil, fmt.Errorf("failed to open client config: %w", err)
	}

	revocationStore, err := system.OpenRevocationStore(scope.Scope("revocation"), clientConfig.Root())
	if err != nil {
		return nil, fmt.Errorf("error opening revocation store: %w", err)
	}

	verifier := system.NewUpstreamVerifier(clientConfig.Upstream(), clientConfig.Root(), revocationStore)

	ep, err := rpc.ConnectToServer(context.Background(), clientConfig.ServerAddr(), clientConfig.Credentials(), clientConfig.Upstream(), verifier)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	verifier.SetEndPoint(ep)

	client := &Client{
		profile:      profile,
		clientConfig: clientConfig,
	}

	return client, nil
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

func (c *Client) Devices() util.ObservableMap[string, *rmm.Device] {
	panic("not implemented")
}

func (c *Client) Tunnels() *rmm.TunnelHandler {
	panic("not implemented")
}
