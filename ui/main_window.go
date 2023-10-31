package ui

import (
	"context"
	"rahnit-rmm/pki"
	"rahnit-rmm/rpc"
	managment "rahnit-rmm/ui/device_managment"
	"rahnit-rmm/ui/enrollment"
	"rahnit-rmm/ui/mainview.go"

	"fyne.io/fyne/v2"
)

func startMainMenu(window fyne.Window, credentials *pki.PermanentCredentials) {
	m := mainview.NewMainView()

	ep, err := rpc.ConnectToUpstream(context.Background(), credentials)
	if err != nil {
		panic(err)
	}

	manageView := managment.NewDeviceManagementView(m, ep)

	enrollView := enrollment.NewEnrollmentView(m, ep, credentials)

	m.Display(window, []mainview.View{
		manageView,
		enrollView,
	})
}
