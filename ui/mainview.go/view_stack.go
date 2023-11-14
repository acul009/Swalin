package mainview

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Widget = (*ViewStack)(nil)

type ViewStack struct {
	widget.BaseWidget
	stack []fyne.CanvasObject
}

func NewViewStack() *ViewStack {
	v := &ViewStack{}
	v.ExtendBaseWidget(v)
	return v
}

func (v *ViewStack) Push(obj fyne.CanvasObject) {
	v.stack = append(v.stack, obj)
	v.Refresh()
}

func (v *ViewStack) Pop() {
	if len(v.stack) <= 1 {
		return
	}
	v.stack[len(v.stack)-1].Hide()
	v.stack = v.stack[:len(v.stack)-1]
	v.stack[len(v.stack)-1].Resize(v.Size())
	v.stack[len(v.stack)-1].Show()
	v.Refresh()
}

func (v *ViewStack) CreateRenderer() fyne.WidgetRenderer {
	return &viewStackRenderer{
		widget: v,
	}
}

type viewStackRenderer struct {
	widget *ViewStack
}

func (v *viewStackRenderer) Layout(size fyne.Size) {
	v.widget.stack[len(v.widget.stack)-1].Resize(size)
}

func (v *viewStackRenderer) MinSize() fyne.Size {
	return fyne.NewSize(500, 300)
}

func (v *viewStackRenderer) Refresh() {
	v.widget.stack[len(v.widget.stack)-1].Refresh()
}

func (v *viewStackRenderer) Destroy() {

}

func (v *viewStackRenderer) Objects() []fyne.CanvasObject {
	if v.widget.stack == nil {
		return []fyne.CanvasObject{}
	}

	if len(v.widget.stack) == 0 {
		return []fyne.CanvasObject{}
	}

	return v.widget.stack[len(v.widget.stack)-1:]
}
