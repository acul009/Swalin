package managment

import (
	"context"
	"rahnit-rmm/rpc"
	"rahnit-rmm/ui/mainview.go"
	"rahnit-rmm/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type deviceManagementView struct {
	main       *mainview.MainView
	ep         *rpc.RpcEndpoint
	devices    *util.ObservableMap[string, rpc.DeviceInfo]
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

	return &deviceManagementView{
		main:       main,
		ep:         ep,
		devices:    devices,
		deviceList: list,
		visible:    false,
	}
}

func (m *deviceManagementView) Icon() fyne.Resource {
	return theme.ComputerIcon()
}

func (m *deviceManagementView) Name() string {
	return "Devices"
}

func (m *deviceManagementView) Prepare() fyne.CanvasObject {
	return m.deviceList.Display
}

func (m *deviceManagementView) Close() {

}
