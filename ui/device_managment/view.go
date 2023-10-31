package managment

import (
	"rahnit-rmm/rpc"
	"rahnit-rmm/ui/mainview.go"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type deviceManagementView struct {
	main *mainview.MainView
	ep   *rpc.RpcEndpoint
}

func NewDeviceManagementView(main *mainview.MainView, ep *rpc.RpcEndpoint) *deviceManagementView {
	return &deviceManagementView{
		main: main,
		ep:   ep,
	}
}

func (m *deviceManagementView) Icon() fyne.Resource {
	return theme.ComputerIcon()
}

func (m *deviceManagementView) Name() string {
	return "Devices"
}

func (m *deviceManagementView) Prepare() fyne.CanvasObject {
	return nil
}

func (m *deviceManagementView) Close() {

}
