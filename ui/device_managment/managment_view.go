package managment

import (
	"context"
	"rahnit-rmm/rpc"
	"rahnit-rmm/ui/mainview.go"
	"rahnit-rmm/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Widget = (*deviceManagementView)(nil)

type deviceManagementView struct {
	widget.BaseWidget
	running util.AsyncAction
	main    *mainview.MainView
	ep      *rpc.RpcEndpoint
	devices util.ObservableMap[string, rpc.DeviceInfo]
}

func NewDeviceManagementView(main *mainview.MainView, ep *rpc.RpcEndpoint) *deviceManagementView {
	devices := util.NewObservableMap[string, rpc.DeviceInfo]()

	m := &deviceManagementView{
		main:    main,
		ep:      ep,
		devices: devices,
	}

	m.ExtendBaseWidget(m)

	return m
}

func (m *deviceManagementView) Show() {
	if m.running != nil {
		return
	}

	cmd := rpc.NewGetDevicesCommand(m.devices)

	running, err := m.ep.SendCommand(context.Background(), cmd)
	if err != nil {
		panic(err)
	}

	m.running = running

	go func() {
		err := running.Wait()
		if err != nil {
			panic(err)
		}
	}()
}

func (m *deviceManagementView) Hide() {
	if m.running == nil {
		return
	}

	go func() {
		err := m.running.Close()
		if err != nil {
			panic(err)
		}
	}()
}

func (m *deviceManagementView) Icon() fyne.Resource {
	return theme.ComputerIcon()
}

func (m *deviceManagementView) Name() string {
	return "Devices"
}

func (m *deviceManagementView) CreateRenderer() fyne.WidgetRenderer {
	icon := theme.ComputerIcon()
	onlineIcon := theme.NewSuccessThemedResource(icon)
	offlineIcon := theme.NewErrorThemedResource(icon)

	return &deviceManagmentViewRenderer{
		widget:      m,
		layout:      layout.NewGridLayoutWithColumns(4),
		onlineIcon:  widget.NewIcon(onlineIcon),
		offlineIcon: offlineIcon,
	}
}

type deviceManagmentViewRenderer struct {
	widget       *deviceManagementView
	layout       fyne.Layout
	onlineIcon   *widget.Icon
	offlineIcon  fyne.Resource
	devices      util.ObservableMap[string, rpc.DeviceInfo]
	deviceLabels []*widget.Label
}

func (v *deviceManagmentViewRenderer) Layout(size fyne.Size) {

}

func (v *deviceManagmentViewRenderer) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

func (v *deviceManagmentViewRenderer) Refresh() {

}

func (v *deviceManagmentViewRenderer) Destroy() {

}

func (v *deviceManagmentViewRenderer) Objects() []fyne.CanvasObject {
}
