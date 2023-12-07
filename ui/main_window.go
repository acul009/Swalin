package ui

import (
	"github.com/rahn-it/svalin/system/client"
	managment "github.com/rahn-it/svalin/ui/device_managment"
	"github.com/rahn-it/svalin/ui/mainview.go"
	"github.com/rahn-it/svalin/ui/tunnels"

	"fyne.io/fyne/v2"
)

func startMainMenu(window fyne.Window, client *client.Client) {
	m := mainview.NewMainView()

	manageView := managment.NewDeviceManagementView(m, client)

	tunnelView := tunnels.NewOpenTunnelsView(client)

	// enrollView := enrollment.NewEnrollmentView(m, cli, credentials)

	m.Display(window, []mainview.MenuView{
		manageView,
		tunnelView,
		// enrollView,
	})
}
