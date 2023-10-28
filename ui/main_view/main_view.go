package mainview

import (
	"rahnit-rmm/pki"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func DisplayMainView(w fyne.Window, credentials *pki.PermanentCredentials) {
	w.SetContent(container.NewVBox(
		widget.NewLabel("Ready!"),
	))
}
