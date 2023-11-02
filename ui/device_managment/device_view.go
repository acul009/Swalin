package managment

import (
	"context"
	"rahnit-rmm/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type deviceView struct {
	ep        *rpc.RpcEndpoint
	container *fyne.Container
}

func newDeviceView(ep *rpc.RpcEndpoint, device rpc.DeviceInfo) *deviceView {
	cont := container.NewVBox(
		widget.NewLabel(device.Name()),
		widget.NewLabel(string(device.Certificate.PemEncode())),
		widget.NewButton("Ping", func() {
			cmd := &rpc.PingCmd{}
			err := ep.SendCommandTo(context.Background(), device.Certificate, cmd)
			if err != nil {
				panic(err)
			}
		}),
	)

	return &deviceView{
		ep:        ep,
		container: cont,
	}
}

func (d *deviceView) Prepare() fyne.CanvasObject {
	return d.container
}

func (d *deviceView) Close() {

}
