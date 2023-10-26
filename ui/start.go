package ui

import (
	"fmt"
	"rahnit-rmm/pki"
	"rahnit-rmm/rpc"
	"regexp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/cmd/fyne_demo/data"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

var master fyne.Window

func StartUI() {
	a := app.NewWithID("io.fyne.demo")
	a.SetIcon(data.FyneLogo)
	w := a.NewWindow("Fyne Demo")

	w.SetMaster()

	master = w

	if pki.Root.Available() {
		unlock()
	} else {
		setup()
	}

	w.Resize(fyne.NewSize(640, 460))
	w.ShowAndRun()
}

func unlock() {
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

	master.SetContent(container.NewVBox(
		infoLabel,
		passwordField,
		loginButton,
	))
}

func setup() {

	master.SetContent(
		container.NewAppTabs(
			container.NewTabItem("Setup Server", container.NewVBox(
				setupServerForm(),
			)),
			container.NewTabItem("Login", container.NewVBox(
				setupLoginForm(),
			)),
		),
	)
}

func setupLoginForm() *widget.Form {
	addressBind := binding.NewString()
	addressInput := widget.NewEntryWithData(addressBind)

	userBind := binding.NewString()
	userInput := widget.NewEntryWithData(userBind)

	passwordBind := binding.NewString()
	passwordInput := widget.NewPasswordEntry()
	passwordInput.Bind(passwordBind)
	passwordInput.Validator = func(s string) error {
		if len(s) < 8 {
			return fmt.Errorf("password must be at least 8 characters")
		}

		return nil
	}

	totpCodeBind := binding.NewString()
	totpInput := widget.NewEntryWithData(totpCodeBind)
	totpInput.PlaceHolder = "00000000"
	totpInput.Validator = func(s string) error {
		regexp := regexp.MustCompile(`^[0-9]{8}$`)
		if !regexp.MatchString(s) {
			return fmt.Errorf("invalid TOTP code")
		}

		return nil
	}

	form := widget.NewForm(
		widget.NewFormItem("Address", addressInput),
		widget.NewFormItem("User", userInput),
		widget.NewFormItem("Password", passwordInput),
		widget.NewFormItem("TOTP", totpInput),
	)

	form.OnSubmit = func() {
		addr, err := addressBind.Get()
		if err != nil {
			return
		}

		username, err := userBind.Get()
		if err != nil {
			return
		}

		password, err := passwordBind.Get()
		if err != nil {
			return
		}

		totpCode, err := totpCodeBind.Get()
		if err != nil {
			return
		}

		err = rpc.Login(addr, username, []byte(password), totpCode)
		if err != nil {
			fmt.Println(err)
		}
	}

	return form
}

func setupServerForm() *widget.Form {
	addressBind := binding.NewString()
	addressInput := widget.NewEntryWithData(addressBind)

	userBind := binding.NewString()
	userInput := widget.NewEntryWithData(userBind)

	passwordBind := binding.NewString()
	passwordInput := widget.NewPasswordEntry()
	passwordInput.Bind(passwordBind)
	passwordInput.Validator = func(s string) error {
		if len(s) < 8 {
			return fmt.Errorf("password must be at least 8 characters")
		}

		return nil
	}

	passwordRepeatBind := binding.NewString()
	passwordRepeatInput := widget.NewPasswordEntry()
	passwordRepeatInput.Bind(passwordRepeatBind)
	passwordRepeatInput.Validator = func(s string) error {
		password, err := passwordBind.Get()
		if err != nil {
			return err
		}

		if password != s {
			return fmt.Errorf("passwords do not match")
		}

		return nil
	}

	totpCodeBind := binding.NewString()
	totpInput := widget.NewEntryWithData(totpCodeBind)
	totpInput.PlaceHolder = "00000000"
	totpInput.Validator = func(s string) error {
		regexp := regexp.MustCompile(`^[0-9]{8}$`)
		if !regexp.MatchString(s) {
			return fmt.Errorf("invalid TOTP code")
		}

		return nil
	}

	form := widget.NewForm(
		widget.NewFormItem("Address", addressInput),
		widget.NewFormItem("Root User", userInput),
		widget.NewFormItem("Password", passwordInput),
		widget.NewFormItem("Repeat Password", passwordRepeatInput),
		widget.NewFormItem("TOTP", totpInput),
	)

	form.OnSubmit = func() {
		addr, err := addressBind.Get()
		if err != nil {
			return
		}

		username, err := userBind.Get()
		if err != nil {
			return
		}

		password, err := passwordBind.Get()
		if err != nil {
			return
		}

		totpCode, err := totpCodeBind.Get()
		if err != nil {
			return
		}

		err = rpc.Login(addr, username, []byte(password), totpCode)
		if err != nil {
			fmt.Println(err)
		}
	}

	return form
}
