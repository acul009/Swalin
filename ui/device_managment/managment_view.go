package managment

import (
	"log"

	"github.com/rahn-it/svalin/rmm"
	"github.com/rahn-it/svalin/system/client"
	"github.com/rahn-it/svalin/ui/components"
	"github.com/rahn-it/svalin/ui/mainview.go"
	"github.com/rahn-it/svalin/util"

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
	cli     *client.Client
}

func NewDeviceManagementView(main *mainview.MainView, cli *client.Client) *deviceManagementView {

	m := &deviceManagementView{
		main: main,
		cli:  cli,
	}

	m.ExtendBaseWidget(m)

	return m
}

func (m *deviceManagementView) Hide() {
	defer m.BaseWidget.Hide()

	log.Printf("Hiding device management view")

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

	table := components.NewTable(m.cli.Devices(),
		components.Column(
			func() *widget.Icon {
				return widget.NewIcon(offlineIcon)
			},
			func(device *rmm.Device, icon *widget.Icon) {
				if device.DeviceInfo.LiveInfo.Online {
					icon.SetResource(onlineIcon)
				} else {
					icon.SetResource(offlineIcon)
				}
				icon.Refresh()
			},
		),
		components.NamedColumn(
			"Name",
			func() *widget.Label {
				return widget.NewLabel("Name")
			},
			func(device *rmm.Device, label *widget.Label) {
				label.SetText(device.Name())
				label.Refresh()
			},
		),
		components.Column(
			func() *layout.Spacer {
				return layout.NewSpacer().(*layout.Spacer)
			},
			func(device *rmm.Device, spacer *layout.Spacer) {
			},
		),
		components.Column(
			func() *widget.Button {
				return widget.NewButton("Connect", func() {})
			},
			func(device *rmm.Device, button *widget.Button) {
				button.OnTapped = func() {
					m.main.PushView(newDeviceView(m.cli, m.main, device))
				}
			},
		),
	)

	log.Printf("Creating device management view renderer")

	return &deviceManagmentViewRenderer{
		widget:    m,
		table:     table,
		testLabel: widget.NewLabel("test"),
	}
}

type deviceManagmentViewRenderer struct {
	widget    *deviceManagementView
	table     *components.Table[string, *rmm.Device]
	testLabel *widget.Label
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
	return []fyne.CanvasObject{v.testLabel, v.table}
}
