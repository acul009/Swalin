package managment

import (
	"fmt"
	"strconv"

	"github.com/rahn-it/svalin/rmm"
	"github.com/rahn-it/svalin/system/client"
	"github.com/rahn-it/svalin/ui/components"
	"github.com/rahn-it/svalin/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Widget = (*tunnelDisplay)(nil)

type tunnelDisplay struct {
	widget.BaseWidget
	cli        *client.Client
	device     *rmm.Device
	config     util.Observable[*rmm.TunnelConfig]
	tcpList    *fyne.Container
	tcpAdd     *fyne.Container
	tcpTunnels util.UpdateableMap[string, *rmm.TcpTunnel]
}

func newTunnelDisplay(cli *client.Client, device *rmm.Device) *tunnelDisplay {
	d := &tunnelDisplay{
		cli:    cli,
		device: device,
		config: device.TunnelConfig(),
	}

	d.ExtendBaseWidget(d)

	d.tcpTunnels = util.NewSyncedMap[string, *rmm.TcpTunnel](
		func(m util.UpdateableMap[string, *rmm.TcpTunnel]) {
			d.config.Subscribe(func(tc *rmm.TunnelConfig) {
				if tc == nil {
					return
				}

				// TODO
			})
		},
		func(m util.UpdateableMap[string, *rmm.TcpTunnel]) {},
	)

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

		fmt.Printf("%+v\n", tunnel)

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
		scrollContainer: container.NewVScroll(
			container.NewVBox(
				components.NewTable(util.ObservableMap[string, *rmm.TcpTunnel](t.tcpTunnels),
					components.NamedColumn(
						"Name",
						func() *widget.Label {
							return widget.NewLabel("Name")
						},
						func(tunnel *rmm.TcpTunnel, label *widget.Label) {
							label.SetText(tunnel.Name)
							label.Refresh()
						},
					),
					components.NamedColumn(
						"Listen Port",
						func() *widget.Label {
							return widget.NewLabel("Listen Port")
						},
						func(tunnel *rmm.TcpTunnel, label *widget.Label) {
							label.SetText(fmt.Sprintf("%d", tunnel.ListenPort))
							label.Refresh()
						},
					),
					components.NamedColumn(
						"Target",
						func() *widget.Label {
							return widget.NewLabel("Target")
						},
						func(tunnel *rmm.TcpTunnel, label *widget.Label) {
							label.SetText(tunnel.Target)
							label.Refresh()
						},
					),
				),
			),
		),
	}
}

type tunnelDisplayRenderer struct {
	widget          *tunnelDisplay
	scrollContainer *container.Scroll
}

func (t *tunnelDisplayRenderer) Destroy() {
}

func (t *tunnelDisplayRenderer) Layout(size fyne.Size) {

	t.scrollContainer.Resize(size)
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
		t.scrollContainer,
	}
}
