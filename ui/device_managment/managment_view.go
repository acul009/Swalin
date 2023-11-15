package managment

import (
	"context"
	"rahnit-rmm/rpc"
	"rahnit-rmm/ui/mainview.go"
	"rahnit-rmm/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Widget = (*deviceManagementView)(nil)

type deviceManagementView struct {
	widget.BaseWidget
	main       *mainview.MainView
	ep         *rpc.RpcEndpoint
	devices    util.ObservableMap[string, rpc.DeviceInfo]
	deviceList *deviceList
	visible    bool
}

func NewDeviceManagementView(main *mainview.MainView, ep *rpc.RpcEndpoint) *deviceManagementView {
	list := newDeviceList(main, ep)
	devices := util.NewObservableMap[string, rpc.DeviceInfo]()

	devices.Subscribe(list.Set, list.Remove)

	cmd := rpc.NewGetDevicesCommand(devices)

	running, err := ep.SendCommand(context.Background(), cmd)
	if err != nil {
		panic(err)
	}

	go func() {
		err := running.Wait()
		if err != nil {
			panic(err)
		}
	}()

	m := &deviceManagementView{
		main:       main,
		ep:         ep,
		devices:    devices,
		deviceList: list,
		visible:    false,
	}

	m.ExtendBaseWidget(m)

	return m
}

func (m *deviceManagementView) Icon() fyne.Resource {
	return theme.ComputerIcon()
}

func (m *deviceManagementView) Name() string {
	return "Devices"
}

func (m *deviceManagementView) CreateRenderer() fyne.WidgetRenderer {
	return &deviceViewRenderer{
		widget: m,
	}
}

type deviceViewRenderer struct {
	widget *deviceManagementView
}

func (v *deviceViewRenderer) Layout(size fyne.Size) {

}

func (v *deviceViewRenderer) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

func (v *deviceViewRenderer) Refresh() {

}

func (v *deviceViewRenderer) Destroy() {

}

func (v *deviceViewRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{
		v.widget.deviceList,
	}
}
