package fynecharts

import (
	"log"
	"rahnit-rmm/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/exp/constraints"
)

type BarWidget[T constraints.Integer | constraints.Float] struct {
	widget.BaseWidget
	max         T
	current     T
	formatter   func(T) string
	onDestroy   func()
	text        *canvas.Text
	rectMax     *canvas.Rectangle
	rectCurrent *canvas.Rectangle
}

func NewBarWidget[T constraints.Integer | constraints.Float](current util.Observable[T], max T, formatter func(T) string) *BarWidget[T] {
	widget := &BarWidget[T]{
		formatter: formatter,
		max:       max,
		current:   0,
		text:      canvas.NewText("", theme.ForegroundColor()),
		onDestroy: func() {},
	}

	log.Printf("extending basewidget")
	widget.BaseWidget.ExtendBaseWidget(widget)

	log.Printf("subscribing to observable")
	widget.onDestroy = current.Subscribe(
		func(current T) {
			widget.update(current)
		},
	)

	return widget
}

func (bw *BarWidget[T]) update(current T) {
	bw.current = current
	bw.computeLabel()
}

func (bw *BarWidget[T]) computeLabel() {
	bw.text.Text = bw.formatter(bw.current)
	bw.text.Refresh()
}

func (bw *BarWidget[T]) CreateRenderer() fyne.WidgetRenderer {
	return &barRenderer[T]{
		widget: bw,
	}
}

type barRenderer[T constraints.Integer | constraints.Float] struct {
	widget *BarWidget[T]
}

func (br *barRenderer[T]) Layout(size fyne.Size) {
	br.widget.text.Move(fyne.NewPos(0, 0))
	br.widget.text.Resize(size)
}

func (br *barRenderer[T]) MinSize() fyne.Size {
	return br.widget.text.MinSize()
}

func (br *barRenderer[T]) Refresh() {
	br.widget.text.Refresh()
}

func (br *barRenderer[T]) Destroy() {
	br.widget.onDestroy()
}

func (br *barRenderer[T]) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{br.widget.text}
}
