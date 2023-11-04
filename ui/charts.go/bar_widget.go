package fynecharts

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/exp/constraints"
)

type BarWidget[T constraints.Integer | constraints.Float] struct {
	widget.BaseWidget
	max       T
	current   T
	unit      string
	onDestroy func()
	text      *canvas.Text
}

func NewBarWidget[T constraints.Integer | constraints.Float](unit string) *BarWidget[T] {
	return &BarWidget[T]{
		unit:      unit,
		max:       0,
		current:   0,
		text:      canvas.NewText("", theme.TextColor()),
		onDestroy: func() {},
	}
}

func (bw *BarWidget[T]) Update(current T, max T) {
	log.Printf("Updating bar: %d/%d", current, max)
	bw.current = current
	bw.max = max
	bw.computeLabel()
}

func (bw *BarWidget[T]) computeLabel() {
	bw.text.Text = fmt.Sprintf("%.2f/%.2f %s", bw.current, bw.max, bw.unit)
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
	return fyne.NewSize(200, 100)
}

func (br *barRenderer[T]) Refresh() {
	br.widget.text.Refresh()
}

func (br *barRenderer[T]) Destroy() {

}

func (br *barRenderer[T]) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{br.widget.text}
}
