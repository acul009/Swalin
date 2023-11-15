package mainview

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type MainView struct {
	viewStack     *ViewStack
	mainContainer *fyne.Container
	leftMenu      *fyne.Container
	backButton    *widget.Button
}

type MenuView interface {
	fyne.CanvasObject
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
		viewStack:     NewViewStack(),
		mainContainer: main,
		leftMenu:      leftMenu,
		backButton:    backButton,
	}

	backButton.OnTapped = m.popView
	backButton.Disable()

	return m
}

func (m *MainView) SetView(v fyne.CanvasObject) {
	m.viewStack.Set(v)
}

func (m *MainView) PushView(v fyne.CanvasObject) {
	m.viewStack.Push(v)
}

func (m *MainView) popView() {
	m.viewStack.Pop()
	if m.viewStack.StackSize() < 2 {
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
