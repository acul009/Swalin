package enrollment

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/rahn-it/svalin/rpc"
	"github.com/rahn-it/svalin/system/client"
	"github.com/rahn-it/svalin/ui/components"
	"github.com/rahn-it/svalin/ui/mainview.go"
)

var _ mainview.MenuView = (*enrollmentList)(nil)

type enrollmentList struct {
	widget.BaseWidget
	main *mainview.MainView
	cli  *client.Client
}

func (e *enrollmentList) Icon() fyne.Resource {
	return theme.DocumentCreateIcon()
}

func (e *enrollmentList) Name() string {
	return "Enrollments"
}

func NewEnrollmentList(main *mainview.MainView, cli *client.Client) *enrollmentList {

	e := &enrollmentList{
		main: main,
		cli:  cli,
	}

	e.ExtendBaseWidget(e)

	return e
}

func (e *enrollmentList) CreateRenderer() fyne.WidgetRenderer {
	table := components.NewTable[string, *rpc.Enrollment](
		e.cli.Enrollments(),
		components.Column(
			func() *widget.Label {
				return widget.NewLabel("Address")
			},
			func(en *rpc.Enrollment, label *widget.Label) {
				label.SetText(en.Addr)
			},
		),
		components.Column(
			func() *widget.Button {
				return widget.NewButton("Select", func() {

				})
			},
			func(en *rpc.Enrollment, button *widget.Button) {

				button.OnTapped = func() {
					view := NewEnrollDeviceView(e.main, e.cli, en)
					e.main.PushView(view)
				}
			},
		),
	)

	return &enrollmentListRenderer{
		table: table,
	}
}

type enrollmentListRenderer struct {
	table *components.Table[string, *rpc.Enrollment]
}

func (e *enrollmentListRenderer) Layout(size fyne.Size) {

	e.table.Resize(size)
}

func (e *enrollmentListRenderer) MinSize() fyne.Size {

	return e.table.MinSize()
}

func (e *enrollmentListRenderer) Refresh() {

	e.table.Refresh()
}

func (e *enrollmentListRenderer) Destroy() {

}

func (e *enrollmentListRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{e.table}
}
