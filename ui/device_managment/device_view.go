package managment

import (
	"rahnit-rmm/rpc"
	"rahnit-rmm/ui/mainview.go"
	"rahnit-rmm/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type deviceView struct {
	widget.BaseWidget
	main *mainview.MainView
	ep   *rpc.RpcEndpoint
	device *rpc.DeviceInfo
	tabs *container.AppTabs
}

func newDeviceView(ep *rpc.RpcEndpoint, main *mainview.MainView, device *rpc.DeviceInfo) *deviceView {
	d := &deviceView{
		ep: ep,
		main: main,
		device: device,
	}

	d.ExtendBaseWidget(d)



	tabs := container.NewAppTabs(
		container.NewTabItem("Processes", newProcessList())
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

}

func (d *deviceViewRenderer) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

func (d *deviceViewRenderer) Refresh() {

}

func (d *deviceViewRenderer) Destroy() {

}

func (d *deviceViewRenderer) Objects() []fyne.CanvasObject {

}
