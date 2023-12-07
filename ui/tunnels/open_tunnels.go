package tunnels

import (
	"fmt"

	"github.com/rahn-it/svalin/rmm"
	"github.com/rahn-it/svalin/system/client"
	"github.com/rahn-it/svalin/ui/components"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type openTunnelsView struct {
	widget.BaseWidget
	cli *client.Client
}

func NewOpenTunnelsView(cli *client.Client) *openTunnelsView {
	o := &openTunnelsView{
		cli: cli,
	}

	o.ExtendBaseWidget(o)

	return o
}

func (o *openTunnelsView) Icon() fyne.Resource {
	return theme.BrokenImageIcon()
}

func (o *openTunnelsView) Name() string {
	return "Open Tunnels"
}

func (o *openTunnelsView) CreateRenderer() fyne.WidgetRenderer {

	tcpMap := o.cli.Tunnels().TcpTunnels()

	tcpTable := components.NewTable(tcpMap,
		components.NamedColumn(
			"Name",
			func() *widget.Label {
				return widget.NewLabel("Name")
			},
			func(tunnel *rmm.ActiveTcpTunnel, label *widget.Label) {
				label.SetText(tunnel.Name)
				label.Refresh()
			},
		),
		components.NamedColumn(
			"Port",
			func() *widget.Label {
				return widget.NewLabel("Port")
			},
			func(tunnel *rmm.ActiveTcpTunnel, label *widget.Label) {
				label.SetText(fmt.Sprintf("%d", tunnel.ListenPort))
				label.Refresh()
			},
		),
	)

	r := &openTunnelsListRenderer{
		widget: o,
		container: container.NewVScroll(
			container.NewVBox(
				widget.NewLabel("TCP Tunnels"),
				tcpTable,
			),
		),
	}

	return r
}

type openTunnelsListRenderer struct {
	widget    *openTunnelsView
	container *container.Scroll
}

func (o *openTunnelsListRenderer) Layout(size fyne.Size) {

	o.container.Resize(size)
}

func (o *openTunnelsListRenderer) MinSize() fyne.Size {
	contSite := o.container.MinSize()
	return fyne.NewSize(contSite.Width, 400)
}

func (o *openTunnelsListRenderer) Refresh() {

	o.container.Refresh()
}

func (o *openTunnelsListRenderer) Destroy() {

}

func (o *openTunnelsListRenderer) Objects() []fyne.CanvasObject {

	return []fyne.CanvasObject{
		o.container,
	}
}
