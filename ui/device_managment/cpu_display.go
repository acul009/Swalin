package managment

import (
	"fmt"
	"log"
	"rahnit-rmm/rmm"
	fynecharts "rahnit-rmm/ui/charts.go"
	"rahnit-rmm/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Widget = (*cpuDisplay)(nil)

type cpuDisplay struct {
	widget.BaseWidget
	layout      fyne.Layout
	bars        []*fynecharts.BarWidget[float64]
	cores       int
	observable  util.Observable[*rmm.CpuStats]
	unsubscribe []func()
}

func newCpuDisplay(observable util.Observable[*rmm.CpuStats]) *cpuDisplay {
	d := &cpuDisplay{
		layout:     layout.NewHBoxLayout(),
		observable: observable,
	}

	d.ExtendBaseWidget(d)

	d.unsubscribe = []func(){observable.Subscribe(
		func(cpu *rmm.CpuStats) {
			// log.Printf("cores: %d", len(cpu.Usage))
			d.cores = len(cpu.Usage)
			d.fixBars()
		},
	)}

	return d
}

func (d *cpuDisplay) fixBars() {
	// log.Printf("fixing bars...")
	if d.bars == nil {
		d.bars = make([]*fynecharts.BarWidget[float64], 0, d.cores)
	}

	for len(d.bars) < d.cores {
		i := len(d.bars)
		log.Printf("adding bar for core %d", i)

		coreStat := util.DeriveObservable[*rmm.CpuStats, float64](
			d.observable,
			func(cpu *rmm.CpuStats) float64 {
				return cpu.Usage[i]
			},
		)

		log.Printf("crearing new widget")
		bar := fynecharts.NewBarWidget[float64](coreStat, 100, func(f float64) string {
			return fmt.Sprintf("%.0f%%", f)
		})

		log.Printf("adding to layout")
		d.bars = append(d.bars, bar)
	}

	// log.Printf("fixing bars complete")
}

type cpuDisplayRenderer struct {
	widget *cpuDisplay
}

func (d *cpuDisplay) CreateRenderer() fyne.WidgetRenderer {
	return &cpuDisplayRenderer{
		widget: d,
	}
}

func (d *cpuDisplayRenderer) Layout(size fyne.Size) {
	d.widget.layout.Layout(d.Objects(), size)
}

func (d *cpuDisplayRenderer) MinSize() fyne.Size {
	return d.widget.layout.MinSize(d.Objects())
}

func (d *cpuDisplayRenderer) Refresh() {

}

func (d *cpuDisplayRenderer) Destroy() {

}

func (d *cpuDisplayRenderer) Objects() []fyne.CanvasObject {
	obj := make([]fyne.CanvasObject, 0, len(d.widget.bars))

	for _, bar := range d.widget.bars {
		obj = append(obj, bar)
	}

	return obj
}
