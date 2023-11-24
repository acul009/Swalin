package ui

import (
	"context"
	"rahnit-rmm/pki"
	"rahnit-rmm/rmm"
	managment "rahnit-rmm/ui/device_managment"
	"rahnit-rmm/ui/mainview.go"
	"rahnit-rmm/ui/tunnels"

	"fyne.io/fyne/v2"
)

func startMainMenu(window fyne.Window, credentials *pki.PermanentCredentials) {
	m := mainview.NewMainView()

	cli, err := rmm.ClientConnect(context.Background(), credentials)
	if err != nil {
		panic(err)
	}

	manageView := managment.NewDeviceManagementView(m, cli)

	tunnelView := tunnels.NewOpenTunnelsView(cli)

	// enrollView := enrollment.NewEnrollmentView(m, cli, credentials)

	m.Display(window, []mainview.MenuView{
		manageView,
		tunnelView,
		// enrollView,
	})
}
