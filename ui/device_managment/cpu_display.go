package managment

import (
	"log"
	"rahnit-rmm/rmm"
	fynecharts "rahnit-rmm/ui/charts.go"
	"rahnit-rmm/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type cpuDisplay struct {
	widget.BaseWidget
	bars   []*fynecharts.BarWidget[float64]
	layout fyne.Layout
	fyne.Container
	observable util.Observable[*rmm.CpuStats]
	unsub      func()
}

func newCpuDisplay(cpu util.Observable[*rmm.CpuStats]) *cpuDisplay {
	c := &cpuDisplay{
		bars:       make([]*fynecharts.BarWidget[float64], 0),
		layout:     layout.NewHBoxLayout(),
		observable: cpu,
	}

	c.ExtendBaseWidget(c)

	c.unsub = c.observable.Subscribe(c.Update)

	return c
}

func (c *cpuDisplay) Update(cpu *rmm.CpuStats) {
	log.Printf("updating cpu display")
	want := len(cpu.Usage)
	have := len(c.bars)
	if have < want {
		for i := 0; i < (want - have); i++ {
			bar := fynecharts.NewBarWidget[float64]("")
			c.bars = append(c.bars, bar)
		}
	} else if have > want {
		c.bars = c.bars[:want]
	}

	log.Printf("len(c.bars): %d, len(cpu.Usage): %d", len(c.bars), len(cpu.Usage))

	for i, usage := range cpu.Usage {
		c.bars[i].Update(usage, 1)
	}

}

func (c *cpuDisplay) Show() {
	c.unsub = c.observable.Subscribe(c.Update)
}

func (c *cpuDisplay) Hide() {
	if c.unsub != nil {
		c.unsub()
		c.unsub = nil
	}
}

func (c *cpuDisplay) objects() []fyne.CanvasObject {
	obj := make([]fyne.CanvasObject, 0, len(c.bars))

	for _, bar := range c.bars {
		obj = append(obj, bar)
	}

	return obj
}

type cpuDisplayRenderer struct {
	display *cpuDisplay
}

func (c *cpuDisplay) CreateRenderer() fyne.WidgetRenderer {
	return &cpuDisplayRenderer{
		display: c,
	}
}

func (c *cpuDisplayRenderer) Layout(size fyne.Size) {
	c.display.layout.Layout(c.display.objects(), size)
}

func (c *cpuDisplayRenderer) MinSize() fyne.Size {
	return fyne.NewSize(float32(len(c.display.bars)*200), 100)
}

func (c *cpuDisplayRenderer) Refresh() {

}

func (c *cpuDisplayRenderer) Destroy() {

}

func (c *cpuDisplayRenderer) Objects() []fyne.CanvasObject {
	return c.display.objects()
}
