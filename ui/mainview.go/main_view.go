package mainview

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type MainView struct {
	currentView   View
	currentObject fyne.CanvasObject
	mainContainer *fyne.Container
	leftMenu      *fyne.Container
}

type View interface {
	Name() string
	Icon() fyne.Resource
	Prepare() fyne.CanvasObject
	Close()
}

func NewMainView() *MainView {
	leftMenu := container.NewVBox()

	main := container.NewBorder(
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
	)
	return &MainView{
		mainContainer: main,
		leftMenu:      leftMenu,
	}
}

func (m *MainView) SetView(v View) {
	if m.currentObject != nil {
		m.mainContainer.Remove(m.currentObject)
	}

	if m.currentView != nil {
		m.currentView.Close()
	}
	m.currentObject = v.Prepare()

	m.currentView = v
	m.mainContainer.Objects = append(m.mainContainer.Objects, m.currentObject)
	m.mainContainer.Refresh()
}

func (m *MainView) PushView(v View) {

}

func (m *MainView) Display(w fyne.Window, views []View) {
	m.leftMenu.RemoveAll()
	for _, view := range views {
		v := view
		icon := view.Icon()
		if icon == nil {
			m.leftMenu.Add(widget.NewButton(v.Name(), func() {
				m.SetView(v)
			}))
		} else {
			m.leftMenu.Add(widget.NewButtonWithIcon(v.Name(), icon, func() {
				m.SetView(v)
			}))
		}
	}

	w.SetContent(m.mainContainer)
}
