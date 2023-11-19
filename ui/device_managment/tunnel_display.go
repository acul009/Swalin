package managment

import (
	"context"
	"errors"
	"fmt"
	"rahnit-rmm/rmm"
	"rahnit-rmm/rpc"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Widget = (*tunnelDisplay)(nil)

type tunnelDisplay struct {
	widget.BaseWidget
	ep      *rpc.RpcEndpoint
	layout  fyne.Layout
	config  *rmm.TunnelConfig
	tcpList *fyne.Container
	tcpAdd  *fyne.Container
}

func newTunnelDisplay(ep *rpc.RpcEndpoint, device *rpc.DeviceInfo) *tunnelDisplay {
	d := &tunnelDisplay{
		ep:     ep,
		config: &rmm.TunnelConfig{},
		layout: layout.NewBorderLayout(nil, nil, nil, nil),
	}

	d.ExtendBaseWidget(d)

	tunnelName := widget.NewEntry()

	listenPort := widget.NewEntry()
	listenPort.Validator = func(s string) error {
		// cast to int
		port, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("number expected")
		}

		if port < 0 || port > 65535 {
			return fmt.Errorf("port must be between 0 and 65535")
		}

		return nil
	}

	targetAddress := widget.NewEntry()

	submitButton := widget.NewButton("Add", func() {
		err := listenPort.Validate()
		if err != nil {
			return
		}

		port, err := strconv.Atoi(listenPort.Text)
		if err != nil {
			return
		}

		tunnel := rmm.TcpTunnel{
			Name:       tunnelName.Text,
			ListenPort: uint16(port),
			Target:     targetAddress.Text,
		}

		d.config.Tcp = append(d.config.Tcp, tunnel)
		d.tcpList.Refresh()
	})

	d.tcpAdd = container.NewVBox(
		container.NewGridWithRows(
			2,
			widget.NewLabel("Tunnel Name"),
			tunnelName,
			widget.NewLabel("Listen Port"),
			listenPort,
			widget.NewLabel("Target Address"),
			targetAddress,
			widget.NewLabel(""),
			submitButton,
		),
	)

	d.tcpList = container.NewBorder(nil, d.tcpAdd, nil, nil)

	cmd := rmm.NewGetConfigCommand[*rmm.TunnelConfig](device.Certificate.GetPublicKey())

	running, err := d.ep.SendCommand(context.Background(), cmd)
	if err != nil {
		sErr := &rpc.SessionError{}
		if !errors.As(err, &sErr) {
			panic(err)
		}
		if sErr.Code() != 404 {
			panic(err)
		}

		d.config = &rmm.TunnelConfig{
			Tcp: []rmm.TcpTunnel{},
		}
	}

	go func() {
		if running != nil {
			err := running.Wait()
			if err != nil {
				panic(err)
			}

			d.config = cmd.Config()
		}

		list := widget.NewList(
			func() int {
				return len(d.config.Tcp)
			},
			func() fyne.CanvasObject {
				return widget.NewLabel("tunnel")
			},
			func(i int, o fyne.CanvasObject) {
				tunnel := d.config.Tcp[i]
				o.(*widget.Label).SetText(tunnel.Name)
			},
		)

		d.tcpList.Objects = append(d.tcpList.Objects[:1], list)
	}()

	return d
}

func (t *tunnelDisplay) Show() {
	t.BaseWidget.Show()
}

func (t *tunnelDisplay) Hide() {
	t.BaseWidget.Hide()
}

func (t *tunnelDisplay) Close() error {
	return nil
}

func (t *tunnelDisplay) CreateRenderer() fyne.WidgetRenderer {
	return &tunnelDisplayRenderer{
		widget: t,
	}
}

type tunnelDisplayRenderer struct {
	widget *tunnelDisplay
}

func (t *tunnelDisplayRenderer) Destroy() {
}

func (t *tunnelDisplayRenderer) Layout(size fyne.Size) {
	t.widget.layout.Layout(t.Objects(), size)
}

func (t *tunnelDisplayRenderer) MinSize() fyne.Size {
	return fyne.NewSize(500, 300)
}

func (t *tunnelDisplayRenderer) Refresh() {

}

func (t *tunnelDisplayRenderer) Objects() []fyne.CanvasObject {
	if t.widget.config == nil {
		return []fyne.CanvasObject{}
	}

	return []fyne.CanvasObject{
		t.widget.tcpList,
		t.widget.tcpAdd,
	}
}
