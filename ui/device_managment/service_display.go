package managment

import (
	"github.com/rahn-it/svalin/rmm"
	"github.com/rahn-it/svalin/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type serviceList struct {
	widget.BaseWidget
	onDestroy    func()
	serviceStats *rmm.ServiceStats
	display      fyne.Widget
}

func newServiceList(processes util.Observable[*rmm.ServiceStats]) *serviceList {
	p := &serviceList{
		serviceStats: &rmm.ServiceStats{
			Services: []rmm.ServiceInfo{},
		},
	}

	p.ExtendBaseWidget(p)

	table := widget.NewTableWithHeaders(
		func() (rows int, cols int) {
			return len(p.serviceStats.Services), 4
		},
		func() fyne.CanvasObject {
			return container.NewVBox()
		},
		func(cell widget.TableCellID, o fyne.CanvasObject) {
			switch cell.Col {
			case 0:
				o.(*fyne.Container).Objects = []fyne.CanvasObject{widget.NewLabel(p.serviceStats.Services[cell.Row].Name)}

			case 1:
				o.(*fyne.Container).Objects = []fyne.CanvasObject{}

			case 2:
				o.(*fyne.Container).Objects = []fyne.CanvasObject{}

			case 3:
				o.(*fyne.Container).Objects = []fyne.CanvasObject{}
			}
		},
	)

	p.display = table

	p.onDestroy = processes.Subscribe(
		func(services *rmm.ServiceStats) {
			p.serviceStats = services
			p.display.Refresh()
		},
	)

	return p
}

func (p *serviceList) CreateRenderer() fyne.WidgetRenderer {
	return &serviceListRenderer{
		widget: p,
	}
}

type serviceListRenderer struct {
	widget *serviceList
}

func (pr *serviceListRenderer) Layout(size fyne.Size) {
	pr.widget.display.Resize(size)
}

func (pr *serviceListRenderer) MinSize() fyne.Size {
	return fyne.NewSize(500, 300)
}

func (pr *serviceListRenderer) Refresh() {
	pr.widget.display.Refresh()
}

func (pr *serviceListRenderer) Destroy() {
	pr.widget.onDestroy()
	pr.widget.display.CreateRenderer().Destroy()
}

func (pr *serviceListRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{
		pr.widget.display,
	}
}
