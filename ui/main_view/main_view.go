package mainview

import (
	"context"
	"fmt"
	"log"
	"rahnit-rmm/pki"
	"rahnit-rmm/rpc"
	"rahnit-rmm/util"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func DisplayMainView(w fyne.Window, credentials *pki.PermanentCredentials) {
	mainContainer := container.NewStack()

	setView := func(o fyne.CanvasObject) {
		mainContainer.Objects = []fyne.CanvasObject{o}
		mainContainer.Refresh()
		log.Printf("changed main view")
	}

	ep, err := rpc.ConnectToUpstream(context.Background(), credentials)
	if err != nil {
		panic(err)
	}

	w.SetContent(
		container.NewBorder(
			container.NewVBox(
				widget.NewToolbar(
					widget.NewToolbarSpacer(),
					widget.NewToolbarSeparator(),
					widget.NewToolbarAction(theme.AccountIcon(), func() {
						setView(accountView(credentials))
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
						setView(enrollView(ep))
					}),
				),
				widget.NewSeparator(),
			),
			nil,
			mainContainer,
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

func enrollView(ep *rpc.RpcEndpoint) fyne.CanvasObject {

	enrollments := util.NewObservableMap[string, rpc.Enrollment]()

	mutex := sync.Mutex{}
	needsUpdate := false
	values := enrollments.Values()

	enrollments.Subscribe(
		func(key string, enrollment rpc.Enrollment) {
			mutex.Lock()
			defer mutex.Unlock()
			needsUpdate = true
		},
		func(key string) {
			mutex.Lock()
			defer mutex.Unlock()
			needsUpdate = true
		},
	)

	updateCommand := rpc.NewGetPendingEnrollmentsCommand(enrollments)

	go func() {
		err := ep.SendCommand(context.Background(), updateCommand)
		if err != nil {
			panic(err)
		}
	}()

	list := widget.NewList(
		func() int {
			return len(values)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i int, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(values[i].Addr)
		},
	)

	go func() {
		for {
			time.Sleep(time.Second)
			mutex.Lock()
			if needsUpdate {
				values = enrollments.Values()
				list.Refresh()
				needsUpdate = false
			}
			mutex.Unlock()
		}
	}()

	return list
}
