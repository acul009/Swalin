package mainview

import (
	"context"
	"fmt"
	"rahnit-rmm/pki"
	"rahnit-rmm/rpc"

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
	Prepare() fyne.CanvasObject
	Close()
}

func (m *MainView) SetView(v View) {
	if m.currentView != nil {
		m.currentView.Close()
	}
	m.currentView = v
	m.mainContainer.Objects = []fyne.CanvasObject{v.Prepare()}
	m.mainContainer.Refresh()
}

func DisplayMainView(w fyne.Window, credentials *pki.PermanentCredentials) {
	m := &MainView{
		mainContainer: container.NewStack(),
	}

	ep, err := rpc.ConnectToUpstream(context.Background(), credentials)
	if err != nil {
		panic(err)
	}

	enroll := newEnrollmentView(ep)

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
					widget.NewButtonWithIcon("Manage", theme.ComputerIcon(), func() {

					}),
					widget.NewButtonWithIcon("Enroll", theme.FolderNewIcon(), func() {
						m.SetView(enroll)
					}),
				),
				widget.NewSeparator(),
			),
			nil,
			m.mainContainer,
		),
	)
}

func accountView(credentials *pki.PermanentCredentials) fyne.CanvasObject {

	cert, err := credentials.GetCertificate()
	if err != nil {
		panic(err)
	}

	return container.NewVBox(
		container.NewGridWithColumns(2,
			widget.NewLabel("Name:"), widget.NewLabel(cert.GetName()),
			widget.NewLabel("Serial number:"), widget.NewLabel(fmt.Sprintf("%d", cert.SerialNumber)),
			widget.NewLabel("Valid until:"), widget.NewLabel(cert.NotAfter.Format("2006-01-02 15:04:05")),
		),
	)
}
