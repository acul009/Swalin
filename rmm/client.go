package rmm

import (
	"context"
	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/rpc"
	"github.com/rahn-it/svalin/util"
	"log"
)

type Client struct {
	ep            *rpc.RpcEndpoint
	tunnelHandler *tunnelHandler
	devices       util.UpdateableMap[string, *Device]
}

func ClientConnect(ctx context.Context, credentials *pki.PermanentCredentials) (*Client, error) {
	ep, err := rpc.ConnectToUpstream(ctx, credentials)
	if err != nil {
		return nil, err
	}

	c := &Client{
		ep: ep,
	}

	c.initSyncedDeviceList()

	th := newTunnelHandler(ep)
	c.tunnelHandler = th

	return c, nil
}

func (c *Client) Tunnels() *tunnelHandler {
	return c.tunnelHandler
}

func (c *Client) Devices() util.UpdateableMap[string, *Device] {
	return c.devices
}

func (c *Client) dispatch() rpc.Dispatcher {
	return c.ep
}

func (c *Client) Close() error {
	return c.ep.Close(200, "Shutdown")
}

func (c *Client) initSyncedDeviceList() {

	var dRunning util.AsyncAction

	devicesInfo := util.NewSyncedMap[string, *DeviceInfo](
		func(m util.UpdateableMap[string, *DeviceInfo]) {
			cmd := NewGetDevicesCommand(m)
			running, err := c.dispatch().SendCommand(context.Background(), cmd)
			if err != nil {
				log.Printf("Error subscribing to devices: %v", err)
				return
			}
			dRunning = running
		},
		func(_ util.UpdateableMap[string, *DeviceInfo]) {
			err := dRunning.Close()
			if err != nil {
				log.Printf("Error unsubscribing from devices: %v", err)
			}
		},
	)

	var unsub func()

	devices := util.NewSyncedMap[string, *Device](
		func(m util.UpdateableMap[string, *Device]) {
			unsub = devicesInfo.Subscribe(
				func(s string, di *DeviceInfo) {
					m.Update(s, func(d *Device, found bool) (*Device, bool) {
						if !found {
							d = &Device{
								c: c,
							}
						}

						d.DeviceInfo = di
						return d, true
					})
				},
				func(s string, _ *DeviceInfo) {
					m.Delete(s)
				},
			)
		},
		func(m util.UpdateableMap[string, *Device]) {
			unsub()
		},
	)

	c.devices = devices
}
