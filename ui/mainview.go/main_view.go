package mainview

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type MainView struct {
	currentView   View
	viewStack     []View
	currentObject fyne.CanvasObject
	mainContainer *fyne.Container
	leftMenu      *fyne.Container
	backButton    *widget.Button
}

type View interface {
	Prepare() fyne.CanvasObject
	Close()
}

type MenuView interface {
	View
	Name() string
	Icon() fyne.Resource
}

func NewMainView() *MainView {
	leftMenu := container.NewVBox()

	backButton := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), nil)

	main := container.NewBorder(
		container.NewVBox(
			container.NewHBox(
				backButton,
				widget.NewToolbar(
					widget.NewToolbarSpacer(),
					widget.NewToolbarSeparator(),
					widget.NewToolbarAction(theme.AccountIcon(), func() {
					}),
				),
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
	m := &MainView{
		currentView:   nil,
		viewStack:     []View{},
		mainContainer: main,
		leftMenu:      leftMenu,
		backButton:    backButton,
	}

	backButton.OnTapped = m.popView
	backButton.Disable()

	return m
}

func (m *MainView) SetView(v View) {
	m.viewStack = make([]View, 0)
	m.backButton.Disable()
	m.display(v)
}

func (m *MainView) display(v View) {
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
	if m.currentView != nil {
		m.viewStack = append(m.viewStack, m.currentView)
		m.backButton.Enable()
	}

	m.display(v)
}

func (m *MainView) popView() {
	if len(m.viewStack) > 0 {
		v := m.viewStack[len(m.viewStack)-1]
		m.viewStack = m.viewStack[:len(m.viewStack)-1]
		m.display(v)
	}
	if len(m.viewStack) == 0 {
		m.backButton.Disable()
	} else {
		m.backButton.Enable()
	}
}

func (m *MainView) Display(w fyne.Window, views []MenuView) {
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
