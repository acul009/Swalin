package ui

import (
	"crypto/rand"
	"fmt"
	"rahnit-rmm/config"
	"rahnit-rmm/pki"
	"rahnit-rmm/rpc"
	"regexp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/cmd/fyne_demo/data"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/widget"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

var master fyne.Window

func StartUI() {
	a := app.NewWithID("io.fyne.demo")
	a.SetIcon(data.FyneLogo)
	w := a.NewWindow("Fyne Demo")

	w.SetMaster()

	master = w

	if config.V().GetString("upstream.address") != "" {
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
	addressBind := binding.NewString()
	addressInput := widget.NewEntryWithData(addressBind)
	addressInput.PlaceHolder = "localhost:1234"

	form := widget.NewForm(
		widget.NewFormItem("Server Address", addressInput),
	)

	form.SubmitText = "Connect"

	form.OnSubmit = func() {

		address, err := addressBind.Get()
		if err != nil {
			return
		}

		conn, err := rpc.FirstClientConnect(address)
		if err != nil {
			return
		}

		switch conn.GetProtocol() {
		case rpc.ProtoClientLogin:
			setupLoginForm(conn)

		case rpc.ProtoServerInit:
			setupServerForm(conn)

		default:
			conn.Close(400, "")
			return
		}
	}

	master.SetContent(
		form,
	)
}

func setupLoginForm(conn *rpc.RpcConnection) {

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
		widget.NewFormItem("User", userInput),
		widget.NewFormItem("Password", passwordInput),
		widget.NewFormItem("TOTP", totpInput),
	)

	form.OnSubmit = func() {

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

		err = rpc.Login(conn, username, []byte(password), totpCode)
		if err != nil {
			fmt.Println(err)
		}
	}

	master.SetContent(
		form,
	)
}

func setupServerForm(conn *rpc.RpcConnection) {

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

	form := widget.NewForm(
		widget.NewFormItem("Root User", userInput),
		widget.NewFormItem("Password", passwordInput),
		widget.NewFormItem("Repeat Password", passwordRepeatInput),
	)

	form.OnSubmit = func() {
		totpCode, totpSecret, err := askForNewTotp(userInput.Text, master.Canvas())
		if err != nil {
			return
		}

		fmt.Println(totpCode, totpSecret)
		// username, err := userBind.Get()
		// if err != nil {
		// 	return
		// }

		// password, err := passwordBind.Get()
		// if err != nil {
		// 	return
		// }

		// totpCode, err := totpCodeBind.Get()
		// if err != nil {
		// 	return
		// }
	}

	master.SetContent(
		form,
	)
}

func askForNewTotp(accountName string, targetCanvas fyne.Canvas) (code string, secret string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "rahnit-rmm",
		AccountName: accountName,
		SecretSize:  32,
		Rand:        rand.Reader,
		Period:      30,
		Digits:      otp.DigitsEight,
	})

	if err != nil {
		return "", "", fmt.Errorf("error generating totp: %w", err)
	}

	secret = key.URL()

	image, err := key.Image(400, 400)
	if err != nil {
		return "", "", fmt.Errorf("error generating totp image: %w", err)
	}

	codeBinding := binding.NewString()
	codeInput := widget.NewEntryWithData(codeBinding)
	codeInput.PlaceHolder = "00000000"
	codeInput.Validator = validation.NewRegexp(`^[0-9]{8}$`, "invalid TOTP code")

	form := widget.NewForm(
		widget.NewFormItem("Code", codeInput),
	)

	qrCode := canvas.NewImageFromImage(image)
	qrCode.FillMode = canvas.ImageFillContain
	qrCode.SetMinSize(fyne.NewSize(200, 200))

	secretEntry := &widget.Entry{
		Text:      secret,
		MultiLine: true,
		Wrapping:  fyne.TextWrapBreak,
	}
	secretEntry.OnChanged = func(s string) {
		secretEntry.SetText(secret)
	}

	widget.NewModalPopUp(
		container.NewVBox(
			qrCode,
			secretEntry,
			widget.NewLabel(secret),
			form,
		),
		master.Canvas(),
	).Show()

	return
}
