package fynecharts

import (
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
		formatter:   formatter,
		max:         max,
		current:     0,
		text:        canvas.NewText("", theme.ForegroundColor()),
		onDestroy:   func() {},
		rectMax:     canvas.NewRectangle(theme.WarningColor()),
		rectCurrent: canvas.NewRectangle(theme.ForegroundColor()),
	}

	widget.text.Alignment = fyne.TextAlignCenter

	// log.Printf("extending basewidget")
	widget.BaseWidget.ExtendBaseWidget(widget)

	// log.Printf("subscribing to observable")
	widget.onDestroy = current.Subscribe(
		func(current T) {
			widget.update(current)
		},
	)

	return widget
}

func (bw *BarWidget[T]) update(current T) {
	if current > bw.max {
		bw.current = bw.max
	} else {
		bw.current = current
	}
	bw.updateGraphics()
}

func (bw *BarWidget[T]) updateGraphics() {
	bw.text.Text = bw.formatter(bw.current)
	bw.text.Refresh()
	barSize := bw.rectMax.Size()
	barPos := bw.rectMax.Position()
	currentHeight := barSize.Height * float32(bw.current/bw.max)
	bw.rectCurrent.Move(fyne.NewPos(barPos.X, barSize.Height-currentHeight))
	bw.rectCurrent.Resize(fyne.NewSize(barSize.Width, currentHeight))
	bw.rectCurrent.Refresh()
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
	textSize := br.widget.text.MinSize()
	br.widget.text.Resize(fyne.NewSize(size.Width, textSize.Height))
	barHeight := size.Height - textSize.Height
	br.widget.text.Move(fyne.NewPos(0, barHeight))
	br.widget.rectMax.Move(fyne.NewPos(size.Width/4, 0))
	br.widget.rectMax.Resize(fyne.NewSize(size.Width/2, barHeight))
	br.widget.updateGraphics()
}

func (br *barRenderer[T]) MinSize() fyne.Size {
	return fyne.NewSize(50, 100)
}

func (br *barRenderer[T]) Refresh() {
	br.widget.updateGraphics()
	br.widget.rectMax.Refresh()
}

func (br *barRenderer[T]) Destroy() {
	br.widget.onDestroy()
}

func (br *barRenderer[T]) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{
		br.widget.text,
		// br.widget.rectMax,
		br.widget.rectCurrent,
	}
}
