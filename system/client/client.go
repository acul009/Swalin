package client

import (
	"context"
	"fmt"
	"log"

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
	devices      *util.SyncedMap[string, *rmm.Device]
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

	var dRunning util.AsyncAction

	devicesInfo := util.NewSyncedMap[string, *system.DeviceInfo](
		func(m util.UpdateableMap[string, *system.DeviceInfo]) {
			cmd := rmm.NewGetDevicesCommand(m)
			running, err := ep.SendCommand(context.Background(), cmd)
			if err != nil {
				log.Printf("Error subscribing to devices: %v", err)
				return
			}
			dRunning = running
		},
		func(_ util.UpdateableMap[string, *system.DeviceInfo]) {
			err := dRunning.Close()
			if err != nil {
				log.Printf("Error unsubscribing from devices: %v", err)
			}
		},
	)

	var unsub func()

	devices := util.NewSyncedMap[string, *rmm.Device](
		func(m util.UpdateableMap[string, *rmm.Device]) {
			unsub = devicesInfo.Subscribe(
				func(s string, di *system.DeviceInfo) {
					m.Update(s, func(d *rmm.Device, found bool) (*rmm.Device, bool) {
						if !found {
							d = &rmm.Device{
								Dispatch: ep,
							}
						}

						d.DeviceInfo = di
						return d, true
					})
				},
				func(s string, di *system.DeviceInfo) {
					m.Delete(s)
				},
			)
		},
		func(m util.UpdateableMap[string, *rmm.Device]) {
			unsub()
		},
	)

	client := &Client{
		profile:      profile,
		clientConfig: clientConfig,
		ep:           ep,
		devices:      devices,
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
	return c.devices
}

func (c *Client) Tunnels() *rmm.TunnelHandler {
	panic("not implemented")
}
