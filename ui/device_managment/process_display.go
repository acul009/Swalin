package managment

import (
	"fmt"
	"github.com/rahn-it/svalin/rmm"
	"github.com/rahn-it/svalin/ui/components"
	"github.com/rahn-it/svalin/util"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type processList struct {
	widget.BaseWidget
	cli    *rmm.Client
	device *rmm.Device

	list    *components.Table[int32, *rmm.ProcessInfo]
	running util.AsyncAction
}

func newProcessList(device *rmm.Device) *processList {

	p := &processList{
		device: device,
	}
	p.ExtendBaseWidget(p)

	p.list = components.NewTable[int32, *rmm.ProcessInfo](device.Processes(),
		components.Column(
			func() *widget.Label {
				return widget.NewLabel("PID")
			},
			func(process *rmm.ProcessInfo, label *widget.Label) {
				label.SetText(fmt.Sprintf("%d", process.Pid))
				label.Refresh()
			},
		),
		components.Column(
			func() *widget.Label {
				return widget.NewLabel("Name")
			},
			func(process *rmm.ProcessInfo, label *widget.Label) {
				label.SetText(process.Name)
				label.Refresh()
			},
		),
		components.Column(
			func() *widget.Button {
				return &widget.Button{
					Text: "Kill",
				}
			},
			func(process *rmm.ProcessInfo, button *widget.Button) {
				button.OnTapped = func() {
					go func() {
						err := p.device.KillProcess(process.Pid)
						if err != nil {
							log.Printf("error killing process: %v", err)
						}
					}()
				}
			},
		),
	)

	return p
}

func (p *processList) CreateRenderer() fyne.WidgetRenderer {
	return &processListRenderer{
		widget: p,
		scroll: container.NewScroll(p.list),
	}
}

type processListRenderer struct {
	widget *processList
	scroll *container.Scroll
}

func (pr *processListRenderer) Layout(size fyne.Size) {
	pr.scroll.Resize(size)
}

func (pr *processListRenderer) MinSize() fyne.Size {
	return fyne.NewSize(500, 300)
}

func (pr *processListRenderer) Refresh() {
	pr.scroll.Refresh()
}

func (pr *processListRenderer) Destroy() {
}

func (pr *processListRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{
		pr.scroll,
	}
}
