package fynecharts

import (
	"bytes"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"github.com/wcharczuk/go-chart/v2"
)

type chartWidget struct {
	widget.BaseWidget
	renderer func(newSize fyne.Size) ChartProvider
}

type ChartProvider interface {
	Render(rp chart.RendererProvider, w io.Writer) error
}

func NewChartWidget(renderer func(newSize fyne.Size) ChartProvider) *chartWidget {
	cw := &chartWidget{
		renderer: renderer,
	}

	cw.ExtendBaseWidget(cw)
	return cw
}

func (cw *chartWidget) CreateRenderer() fyne.WidgetRenderer {
	cr := &chartWidgetRenderer{
		widget: cw,
	}

	cr.refreshImage(fyne.NewSize(50, 50))

	return cr
}

type chartWidgetRenderer struct {
	widget   *chartWidget
	renderer func(newSize fyne.Size) ChartProvider
	img      *canvas.Image
}

func (cr *chartWidgetRenderer) Layout(newSize fyne.Size) {
	size := newSize.Max(fyne.NewSize(50, 50))
	cr.refreshImage(size)
	cr.img.Move(fyne.NewPos(0, 0))
	cr.img.Resize(size)
}

func (cr *chartWidgetRenderer) refreshImage(size fyne.Size) {
	buf := new(bytes.Buffer)
	err := cr.renderer(size).Render(chart.PNG, buf)
	if err != nil {
		panic(err)
	}
	cr.img = canvas.NewImageFromReader(buf, "")
}

func (cr *chartWidgetRenderer) MinSize() fyne.Size {
	return fyne.NewSize(50, 50) // Adjust as needed
}

func (cr *chartWidgetRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{cr.img}
}

func (cr *chartWidgetRenderer) Refresh() {
	cr.img.Refresh()
}

func (cr *chartWidgetRenderer) Destroy() {}
