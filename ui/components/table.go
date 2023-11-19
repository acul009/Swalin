package components

import (
	"rahnit-rmm/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type Table[T comparable, U any] struct {
	widget.BaseWidget
	m    util.ObservableMap[T, U]
	cols []col[U]
}

func TableColumn[U any, V fyne.CanvasObject](create func() V, update func(U, V)) col[U] {
	return &tableColumn[U, V]{
		createFunc: create,
		updateFunc: update,
	}
}

type tableColumn[U any, V fyne.CanvasObject] struct {
	createFunc func() V
	updateFunc func(U, V)
}

type col[U any] interface {
	newCell() cell[U]
}

func (c *tableColumn[U, V]) newCell() cell[U] {
	return &tableCell[U, V]{
		columnDef: c,
		obj:       c.createFunc(),
	}
}

func NewTable[T comparable, U any](m util.ObservableMap[T, U], cols ...col[U]) *Table[T, U] {
	t := &Table[T, U]{
		m:    m,
		cols: cols,
	}
	t.ExtendBaseWidget(t)

	return t
}

func (t *Table[T, U]) CreateRenderer() fyne.WidgetRenderer {
	rows := make(map[T][]cell[U])

	tr := &tableRenderer[T, U]{
		widget: t,
		rows:   rows,
		layout: layout.NewGridLayoutWithColumns(len(t.cols)),
	}

	tr.unsubscribe = t.m.Subscribe(
		func(key T, value U) {
			row, ok := rows[key]
			if !ok {
				row = make([]cell[U], 0, len(t.cols))
				for _, col := range t.cols {
					row = append(row, col.newCell())
				}

				rows[key] = row
			}

			for _, cell := range row {
				cell.update(value)
			}

			if !ok {
				tr.Refresh()
			}
		},
		func(t T) {
			delete(rows, t)
		},
	)

	for key, val := range t.m.GetAll() {
		row := make([]cell[U], 0, len(t.cols))
		for _, col := range t.cols {
			row = append(rows[key], col.newCell())
		}

		for _, cell := range row {
			cell.update(val)
		}

		rows[key] = row
	}

	return tr

}

type cell[U any] interface {
	update(U)
	object() fyne.CanvasObject
}

type tableCell[U any, V fyne.CanvasObject] struct {
	columnDef *tableColumn[U, V]
	obj       V
}

func (tc *tableCell[U, V]) update(value U) {
	tc.columnDef.updateFunc(value, tc.obj)
}

func (tc *tableCell[U, V]) object() fyne.CanvasObject {
	return tc.obj
}

type tableRenderer[T comparable, U any] struct {
	widget      *Table[T, U]
	rows        map[T][]cell[U]
	unsubscribe func()
	layout      fyne.Layout
}

func (tr *tableRenderer[T, U]) Layout(size fyne.Size) {
	minSize := tr.layout.MinSize(tr.Objects())
	tr.layout.Layout(tr.Objects(), fyne.Size{Width: size.Width, Height: minSize.Height})
}

func (tr *tableRenderer[T, U]) MinSize() fyne.Size {
	return tr.layout.MinSize(tr.Objects())
}

func (tr *tableRenderer[T, U]) Refresh() {
	tr.Layout(tr.widget.Size())
}

func (tr *tableRenderer[T, U]) Destroy() {
	tr.unsubscribe()
}

func (tr *tableRenderer[T, U]) Objects() []fyne.CanvasObject {
	cells := make([]fyne.CanvasObject, 0, len(tr.rows)*len(tr.widget.cols))

	for _, row := range tr.rows {
		for _, cell := range row {
			cells = append(cells, cell.object())
		}
	}

	return cells
}
