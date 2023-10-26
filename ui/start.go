package ui

import (
	"rahnit-rmm/pki"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/cmd/fyne_demo/data"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var master fyne.Window

func StartUI() {
	a := app.NewWithID("io.fyne.demo")
	a.SetIcon(data.FyneLogo)
	w := a.NewWindow("Fyne Demo")

	w.SetMaster()

	master = w

	login()

	w.Resize(fyne.NewSize(640, 460))
	w.ShowAndRun()
}

func login() {
	infoLabel := widget.NewLabel("Please login")

	passwordField := widget.NewPasswordEntry()

	minLength := 8

	loginCallback := func(s string) {
		if len(s) < minLength {
			return
		}
		err := pki.Unlock([]byte(s))
		if err != nil {
			infoLabel.SetText("Incorrect Password")
			return
		}

		infoLabel.SetText("Login Successful")

	}

	loginButton := widget.NewButton("Login", func() {
		loginCallback(passwordField.Text)
	})
	loginButton.Disable()

	passwordField.OnChanged = func(s string) {
		if len(s) < minLength {
			loginButton.Disable()
		} else {
			loginButton.Enable()
		}
	}

	passwordField.OnSubmitted = func(s string) {
		loginCallback(s)
	}

	popUp := widget.NewModalPopUp(
		container.NewVBox(
			infoLabel,
			passwordField,
			loginButton,
		),
		master.Canvas(),
	)

	popUp.Resize(fyne.NewSize(400, 0))

	popUp.Show()
}
