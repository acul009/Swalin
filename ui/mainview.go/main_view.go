package mainview

import (
	"log"

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

	viewStack := NewViewStack()

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
		viewStack,
	)
	m := &MainView{
		viewStack:     viewStack,
		mainContainer: main,
		leftMenu:      leftMenu,
		backButton:    backButton,
	}

	backButton.OnTapped = m.PopView
	backButton.Disable()

	return m
}

func (m *MainView) SetView(v fyne.CanvasObject) {
	log.Printf("Setting new view")
	m.viewStack.Set(v)
	m.refreshBackButton()
}

func (m *MainView) PushView(v fyne.CanvasObject) {
	m.viewStack.Push(v)
	m.refreshBackButton()
}

func (m *MainView) PopView() {
	m.viewStack.Pop()
	m.refreshBackButton()
}

func (m *MainView) refreshBackButton() {

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
