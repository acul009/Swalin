package ui

import (
	"context"
	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/rmm"
	managment "github.com/rahn-it/svalin/ui/device_managment"
	"github.com/rahn-it/svalin/ui/mainview.go"
	"github.com/rahn-it/svalin/ui/tunnels"

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
