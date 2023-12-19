package enrollment

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/rahn-it/svalin/rpc"
	"github.com/rahn-it/svalin/system/client"
	"github.com/rahn-it/svalin/ui/mainview.go"
)

type enrollDeviceView struct {
	widget.BaseWidget
	main       *mainview.MainView
	cli        *client.Client
	enrollment *rpc.Enrollment
}

func NewEnrollDeviceView(main *mainview.MainView, cli *client.Client, enrollment *rpc.Enrollment) *enrollDeviceView {
	edv := &enrollDeviceView{
		main:       main,
		cli:        cli,
		enrollment: enrollment,
	}

	edv.ExtendBaseWidget(edv)

	return edv
}

func (edv *enrollDeviceView) CreateRenderer() fyne.WidgetRenderer {
	nameInput := widget.NewEntry()
	enrollButton := widget.NewButton("Enroll", func() {
		err := edv.cli.EnrollDevice(edv.enrollment.PublicKey, nameInput.Text)
		if err != nil {
			log.Printf("Error enrolling device: %v", err)
		}
		edv.main.PopView()
	})

	return &enrollDeviceViewRenderer{
		widget: edv,
		container: container.NewVBox(
			widget.NewLabel("Enroll Device"),
			layout.NewSpacer(),
			widget.NewLabel("Device Name"),
			nameInput,
			enrollButton,
			layout.NewSpacer(),
		),
	}
}

type enrollDeviceViewRenderer struct {
	widget    *enrollDeviceView
	container *fyne.Container
}

func (edvr *enrollDeviceViewRenderer) MinSize() fyne.Size {
	return edvr.container.MinSize()
}

func (edvr *enrollDeviceViewRenderer) Layout(size fyne.Size) {
	edvr.container.Resize(size)
}

func (edvr *enrollDeviceViewRenderer) Destroy() {
}

func (edvr *enrollDeviceViewRenderer) Refresh() {
	edvr.container.Refresh()
}

func (edvr *enrollDeviceViewRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{edvr.container}
}
