package managment

import (
	"io"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/fyne-io/terminal"
	"github.com/rahn-it/svalin/rmm"
)

type deviceBasicInfo struct {
	widget.BaseWidget
	device *rmm.Device
}

func newDeviceBasicInfo(device *rmm.Device) *deviceBasicInfo {
	return &deviceBasicInfo{
		device: device,
	}
}

func (d *deviceBasicInfo) CreateRenderer() fyne.WidgetRenderer {

	return &deviceBasicInfoRenderer{
		widget: d,
		container: container.NewVBox(
			widget.NewLabel(d.device.Name()),
			container.NewGridWithColumns(2),
			widget.NewButton("Terminal", func() {
				term := terminal.New()
				window := fyne.CurrentApp().NewWindow("Terminal for " + d.device.Name())
				window.Resize(fyne.NewSize(800, 600))
				window.SetContent(term)

				readStdin, writeStdin := io.Pipe()
				readStdout, writeStdout := io.Pipe()

				async, err := d.device.OpenShell(readStdin, writeStdout)
				if err != nil {
					log.Printf("error opening shell: %v", err)
					return
				}

				window.SetOnClosed(func() {
					async.Close()
				})

				go func() {
					err := term.RunWithConnection(writeStdin, readStdout)
					if err != nil {
						log.Printf("error running shell: %v", err)
					}
				}()

				go func() {
					async.Wait()
					window.Close()
				}()

				go func() {
					window.Show()
				}()
			}),
		),
	}
}

type deviceBasicInfoRenderer struct {
	widget    *deviceBasicInfo
	container *fyne.Container
}

func (d *deviceBasicInfoRenderer) MinSize() fyne.Size {
	return d.container.MinSize()
}

func (d *deviceBasicInfoRenderer) Layout(size fyne.Size) {

	d.container.Resize(size)
}

func (d *deviceBasicInfoRenderer) Destroy() {

}

func (d *deviceBasicInfoRenderer) Refresh() {
	d.container.Refresh()
}

func (d *deviceBasicInfoRenderer) Objects() []fyne.CanvasObject {

	return []fyne.CanvasObject{d.container}
}
