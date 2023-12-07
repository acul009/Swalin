package managment

import (
	"github.com/rahn-it/svalin/rmm"
	"github.com/rahn-it/svalin/system/client"
	"github.com/rahn-it/svalin/ui/mainview.go"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type deviceView struct {
	widget.BaseWidget
	main   *mainview.MainView
	cli    *client.Client
	device *rmm.Device
	tabs   *container.AppTabs
}

func newDeviceView(cli *client.Client, main *mainview.MainView, device *rmm.Device) *deviceView {
	d := &deviceView{
		cli:    cli,
		main:   main,
		device: device,
	}

	d.ExtendBaseWidget(d)

	d.tabs = container.NewAppTabs(
		container.NewTabItem("Processes", newProcessList(d.device)),
		container.NewTabItem("Tunnels", newTunnelDisplay(cli, d.device)),
	)

	return d
}

func (d *deviceView) CreateRenderer() fyne.WidgetRenderer {
	return &deviceViewRenderer{
		widget: d,
	}
}

type deviceViewRenderer struct {
	widget *deviceView
}

func (d *deviceViewRenderer) Layout(size fyne.Size) {

	d.widget.tabs.Resize(size)
}

func (d *deviceViewRenderer) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

func (d *deviceViewRenderer) Refresh() {
	d.widget.tabs.Refresh()
}

func (d *deviceViewRenderer) Destroy() {

}

func (d *deviceViewRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{
		d.widget.tabs,
	}
}
