package mainview

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type MainView struct {
	currentView   View
	mainContainer *fyne.Container
}

type View interface {
	Name() string
	Icon() fyne.Resource
	Prepare() fyne.CanvasObject
	Close()
}

func NewMainView() *MainView {
	return &MainView{
		mainContainer: container.NewVBox(),
	}
}

func (m *MainView) SetView(v View) {
	if m.currentView != nil {
		m.currentView.Close()
	}
	m.currentView = v
	m.mainContainer.Objects = []fyne.CanvasObject{v.Prepare()}
	m.mainContainer.Refresh()
}

func (m *MainView) PushView(v View) {

}

func (m *MainView) Display(w fyne.Window, views []View) {

	leftMenu := container.NewVBox()
	for _, view := range views {
		icon := view.Icon()
		if icon == nil {
			leftMenu.Add(widget.NewButton(view.Name(), func() {
				m.SetView(view)
			}))
		} else {
			leftMenu.Add(widget.NewButtonWithIcon(view.Name(), icon, func() {
				m.SetView(view)
			}))
		}
	}

	w.SetContent(
		container.NewBorder(
			container.NewVBox(
				widget.NewToolbar(
					widget.NewToolbarSpacer(),
					widget.NewToolbarSeparator(),
					widget.NewToolbarAction(theme.AccountIcon(), func() {

					}),
				),
				widget.NewSeparator(),
			),
			nil,
			container.NewHBox(
				container.NewVBox(
					leftMenu,
				),
				widget.NewSeparator(),
			),
			nil,
			m.mainContainer,
		),
	)
}
