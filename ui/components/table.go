package components

import (
	"log"
	"rahnit-rmm/util"
	"sync"

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

	tr := &tableRenderer[T, U]{
		widget:      t,
		layout:      layout.NewGridLayoutWithColumns(len(t.cols)),
		mutex:       sync.Mutex{},
		deletedRows: map[int]struct{}{},
	}

	tr.unsubscribe = t.m.Subscribe(
		func(key T, value U) {
			tr.mutex.Lock()
			rowIndex, ok := tr.rowMap[key]
			if !ok {
				row := make([]cell[U], 0, len(t.cols))
				for _, col := range t.cols {
					row = append(row, col.newCell())
				}

				rowIndex = len(tr.cells)

				tr.rowMap[key] = rowIndex
				tr.cells = append(tr.cells, row...)

				log.Printf("adding row for %v", key)
			}

			for _, cell := range tr.cells[rowIndex : rowIndex+len(t.cols)] {
				cell.update(value)
			}

			tr.mutex.Unlock()

			if !ok {
				tr.Refresh()
			}
		},
		func(t T) {
			log.Printf("deleting row for %v", t)
			tr.mutex.Lock()
			defer tr.mutex.Unlock()
			rowIndex, ok := tr.rowMap[t]
			if ok {
				delete(tr.rowMap, t)
				tr.deletedRows[rowIndex] = struct{}{}
			}
		},
	)

	all := t.m.GetAll()

	tr.rowMap = make(map[T]int, len(all))
	tr.cells = make([]cell[U], 0, len(all)*len(t.cols))

	for key, val := range t.m.GetAll() {

		tr.rowMap[key] = len(tr.cells)

		for _, col := range t.cols {
			cell := col.newCell()
			tr.cells = append(tr.cells, cell)
			cell.update(val)
		}
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
	rowMap      map[T]int
	cells       []cell[U]
	unsubscribe func()
	layout      fyne.Layout
	mutex       sync.Mutex
	deletedRows map[int]struct{}
	copy        []fyne.CanvasObject
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
	tr.mutex.Lock()
	defer tr.mutex.Unlock()

	tr.copy = tr.copy[:0]

	for index := 0; index < len(tr.cells); index += len(tr.widget.cols) {
		_, deleted := tr.deletedRows[index]
		if deleted {
			continue
		}

		for offset := 0; offset < len(tr.widget.cols); offset++ {
			tr.copy = append(tr.copy, tr.cells[index+offset].object())
		}
	}

	return tr.copy
}
