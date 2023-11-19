package managment

import (
	"fmt"
	"rahnit-rmm/rmm"
	"rahnit-rmm/rpc"
	"rahnit-rmm/ui/components"
	"rahnit-rmm/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type processList struct {
	widget.BaseWidget
	onDestroy func()
	processes util.ObservableMap[int32, *rmm.ProcessInfo]
	running   util.AsyncAction
}

func newProcessList(ep *rpc.RpcEndpoint, device *rpc.DeviceInfo) *processList {

	p := &processList{
		processes: util.NewObservableMap[int32, *rmm.ProcessInfo](),
	}
	p.ExtendBaseWidget(p)

	components.NewTable[int32, *rmm.ProcessInfo](p.processes,
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
	)

	return p
}

func (p *processList) Show() {
	if p.running != nil {
		return
	}

}

func (p *processList) Hide() {

}

func (p *processList) CreateRenderer() fyne.WidgetRenderer {
	return &processListRenderer{
		widget: p,
	}
}

type processListRenderer struct {
	widget *processList
}

func (pr *processListRenderer) Layout(size fyne.Size) {
	pr.widget.list.Resize(size)
}

func (pr *processListRenderer) MinSize() fyne.Size {
	return fyne.NewSize(500, 300)
}

func (pr *processListRenderer) Refresh() {
	pr.widget.list.Refresh()
}

func (pr *processListRenderer) Destroy() {
	pr.widget.onDestroy()
	pr.widget.list.CreateRenderer().Destroy()
}

func (pr *processListRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{
		pr.widget.list,
	}
}
