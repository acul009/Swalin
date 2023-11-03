package managment

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type cpuDisplay struct {
	mainContainer *fyne.Container
	labels        []*widget.Label
}

func newCpuDisplay() *cpuDisplay {
	return &cpuDisplay{
		mainContainer: container.NewHBox(),
	}
}

func (c *cpuDisplay) canvasObject() *fyne.Container {
	return c.mainContainer
}

func (c *cpuDisplay) update(cpu []float64) {
	if c.labels == nil {
		c.labels = make([]*widget.Label, len(cpu))
		for i := range c.labels {
			c.labels[i] = widget.NewLabel("")
			c.mainContainer.Add(c.labels[i])
		}
	}

	for i, v := range cpu {
		c.labels[i].SetText(fmt.Sprintf("%s %%", strconv.FormatFloat(v, 'f', 0, 64)))
	}

}
