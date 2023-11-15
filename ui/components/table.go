package components

import (
	"rahnit-rmm/util"

	"fyne.io/fyne"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type Table[T comparable, U any] struct {
	widget.BaseWidget
	m util.ObservableMap[T, U]
}

type tableColumn[U any, V fyne.CanvasObject] struct {
	Create func() V
	Update func(U, V)
}

func NewTable[T comparable, U any](util.ObservableMap[T, U]) *Table[T, U] {
	t := &Table[T, U]{}
	t.ExtendBaseWidget(t)

	return t
}

func (t *Table[T, U]) CreateRenderer() fyne.WidgetRenderer {

}

type cell[U any] interface {
	update(U)
	object() fyne.CanvasObject
}

type tableCell[U any, V fyne.CanvasObject] struct {
	columnDef tableColumn[U, V]
	obj       V
}

func (tc *tableCell[U, V]) update(value U) {
	tc.columnDef.Update(value, tc.object)
}

func (tc *tableCell[U, V]) object() fyne.CanvasObject {
	return tc.obj
}

func newTableCell[U any, V fyne.CanvasObject](columnDef tableColumn[U, V]) *tableCell[U, V] {
	return &tableCell[U, V]{
		columnDef: columnDef,
	}
}

type tableRenderer[T comparable, U any] struct {
	widget *Table[T, U]
	rows   map[T][]cell[U]
}

func (tr *tableRenderer[T, U]) Layout(size fyne.Size) {

}

func (tr *tableRenderer[T, U]) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

func (tr *tableRenderer[T, U]) Refresh() {

}

func (tr *tableRenderer[T, U]) Destroy() {

}

func (tr *tableRenderer[T, U]) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{}
}
