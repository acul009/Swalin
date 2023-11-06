package managment

import (
	"fmt"
	"rahnit-rmm/rmm"
	"rahnit-rmm/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type processList struct {
	widget.BaseWidget
	onDestroy    func()
	processStats *rmm.ProcessStats
	list         *widget.List
}

func newProcessList(processes util.Observable[*rmm.ProcessStats]) *processList {
	p := &processList{
		processStats: &rmm.ProcessStats{
			Processes: []rmm.ProcessInfo{},
		},
	}

	p.list = widget.NewList(
		func() int {
			return len(p.processStats.Processes)
		},
		func() fyne.CanvasObject {
			pid := widget.NewLabel("PID   ")
			name := widget.NewLabel("Name")

			return container.NewHBox(
				pid,
				name,
			)
		},
		func(i int, o fyne.CanvasObject) {
			pid := o.(*fyne.Container).Objects[0].(*widget.Label)
			name := o.(*fyne.Container).Objects[1].(*widget.Label)
			processInfo := p.processStats.Processes[i]
			pidString := fmt.Sprintf("%d", processInfo.Pid)
			pid.Text = fmt.Sprintf("%*s", 6-len(pidString), pidString)
			name.Text = processInfo.Name
			o.Refresh()
		},
	)

	p.ExtendBaseWidget(p)

	p.onDestroy = processes.Subscribe(
		func(processes *rmm.ProcessStats) {
			p.processStats = processes
			p.list.Refresh()
		},
	)

	return p
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
