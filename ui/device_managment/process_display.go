package managment

import (
	"context"
	"fmt"
	"log"
	"rahnit-rmm/rmm"
	"rahnit-rmm/rpc"
	"rahnit-rmm/ui/components"
	"rahnit-rmm/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type processList struct {
	widget.BaseWidget
	ep        *rpc.RpcEndpoint
	device    *rpc.DeviceInfo
	processes util.ObservableMap[int32, *rmm.ProcessInfo]
	list      *components.Table[int32, *rmm.ProcessInfo]
	running   util.AsyncAction
}

func newProcessList(ep *rpc.RpcEndpoint, device *rpc.DeviceInfo) *processList {

	p := &processList{
		ep:        ep,
		device:    device,
		processes: util.NewObservableMap[int32, *rmm.ProcessInfo](),
	}
	p.ExtendBaseWidget(p)

	p.list = components.NewTable[int32, *rmm.ProcessInfo](p.processes,
		components.TableColumn(
			func() *widget.Label {
				return widget.NewLabel("PID")
			},
			func(process *rmm.ProcessInfo, label *widget.Label) {
				label.SetText(fmt.Sprintf("%d", process.Pid))
				label.Refresh()
			},
		),
		components.TableColumn(
			func() *widget.Label {
				return widget.NewLabel("Name")
			},
			func(process *rmm.ProcessInfo, label *widget.Label) {
				label.SetText(process.Name)
				label.Refresh()
			},
		),
		components.TableColumn(
			func() *widget.Button {
				return &widget.Button{
					Text: "Kill",
				}
			},
			func(process *rmm.ProcessInfo, button *widget.Button) {
				button.OnTapped = func() {
					running, err := p.ep.SendCommandTo(context.Background(), p.device.Certificate, rmm.NewKillProcessCommand(process.Pid))
					if err != nil {
						log.Printf("error running command: %v", err)
					}
					go func() {
						err := running.Wait()
						if err != nil {
							log.Printf("error running command: %v", err)
						}
					}()
				}
			},
		),
	)

	return p
}

func (p *processList) Show() {
	if p.running == nil {

		cmd := rmm.NewMonitorProcessesCommand(p.processes)

		running, err := p.ep.SendCommandTo(context.Background(), p.device.Certificate, cmd)
		if err != nil {
			log.Printf("error running command: %v", err)
		}

		p.running = running
	}

	p.BaseWidget.Show()
}

func (p *processList) Hide() {
	if p.running != nil {
		p.running.Close()
	}

	p.BaseWidget.Hide()
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
