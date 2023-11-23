package mainview

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Widget = (*ViewStack)(nil)

type ViewStack struct {
	widget.BaseWidget
	stack []fyne.CanvasObject
}

func NewViewStack() *ViewStack {
	v := &ViewStack{
		stack: make([]fyne.CanvasObject, 0, 5),
	}
	v.ExtendBaseWidget(v)
	return v
}

func (v *ViewStack) Push(obj fyne.CanvasObject) {
	if len(v.stack) > 0 {
		v.stack[len(v.stack)-1].Hide()
	}
	v.stack = append(v.stack, obj)
	v.showTop()
}

func (v *ViewStack) Set(obj fyne.CanvasObject) {
	if len(v.stack) > 0 {
		v.stack[len(v.stack)-1].Hide()
	}
	v.stack = v.stack[:0]
	v.stack = append(v.stack, obj)
	v.showTop()
}

func (v *ViewStack) Pop() {
	log.Printf("Popping view stack")
	if len(v.stack) <= 1 {
		return
	}
	v.stack[len(v.stack)-1].Hide()
	v.stack = v.stack[:len(v.stack)-1]
	v.showTop()
}

func (v *ViewStack) StackSize() int {
	return len(v.stack)
}

func (v *ViewStack) showTop() {
	if len(v.stack) == 0 {
		return
	}

	log.Printf("view stack size: %v", v.Size())

	top := v.stack[len(v.stack)-1]
	top.Resize(v.Size())
	top.Show()
	top.Refresh()
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
	if len(v.widget.stack) == 0 {
		return
	}
	log.Printf("ViewStack size: %v", size)
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
