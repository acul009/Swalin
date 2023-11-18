package managment

import (
	"context"
	"log"
	"rahnit-rmm/rpc"
	"rahnit-rmm/ui/components"
	"rahnit-rmm/ui/mainview.go"
	"rahnit-rmm/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Widget = (*deviceManagementView)(nil)

type deviceManagementView struct {
	widget.BaseWidget
	running util.AsyncAction
	main    *mainview.MainView
	ep      *rpc.RpcEndpoint
	devices util.ObservableMap[string, *rpc.DeviceInfo]
}

func NewDeviceManagementView(main *mainview.MainView, ep *rpc.RpcEndpoint) *deviceManagementView {
	devices := util.NewObservableMap[string, *rpc.DeviceInfo]()

	m := &deviceManagementView{
		main:    main,
		ep:      ep,
		devices: devices,
	}

	m.ExtendBaseWidget(m)

	return m
}

func (m *deviceManagementView) Show() {
	log.Printf("Showing device management view")

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

	m.BaseWidget.Show()
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

	m.BaseWidget.Hide()
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

	table := components.NewTable[string, *rpc.DeviceInfo](m.devices,
		components.TableColumn(
			func() *widget.Icon {
				return widget.NewIcon(offlineIcon)
			},
			func(device *rpc.DeviceInfo, icon *widget.Icon) {
				if device.Online {
					icon.SetResource(onlineIcon)
				} else {
					icon.SetResource(offlineIcon)
				}
				icon.Refresh()
			},
		),
		components.TableColumn(
			func() *widget.Label {
				return widget.NewLabel("Name")
			},
			func(device *rpc.DeviceInfo, label *widget.Label) {
				label.SetText(device.Name())
				label.Refresh()
			},
		),
	)

	log.Printf("Creating device management view renderer")

	return &deviceManagmentViewRenderer{
		widget: m,
		table:  table,
	}
}

type deviceManagmentViewRenderer struct {
	widget *deviceManagementView
	table  *components.Table[string, *rpc.DeviceInfo]
}

func (v *deviceManagmentViewRenderer) Layout(size fyne.Size) {

	v.table.Resize(size)
}

func (v *deviceManagmentViewRenderer) MinSize() fyne.Size {
	return fyne.NewSize(400, 400)
}

func (v *deviceManagmentViewRenderer) Refresh() {
	log.Printf("Refreshing device management view")
	v.table.Refresh()
}

func (v *deviceManagmentViewRenderer) Destroy() {

}

func (v *deviceManagmentViewRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{v.table}
}
