package ui

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/rahn-it/svalin/config"
	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/rpc"
	"github.com/rahn-it/svalin/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/cmd/fyne_demo/data"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func StartUI() {

	a := app.NewWithID("io.fyne.demo")
	a.SetIcon(data.FyneLogo)
	w := a.NewWindow("Fyne Demo")

	if config.V().GetString("upstream.address") != "" {
		unlock(w)
	} else {
		setup(w)
	}

	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}

func unlock(w fyne.Window) {
	infoLabel := widget.NewLabel("Please login")

	availableUsers, err := pki.ListAvailableUserCredentials()
	if err != nil {
		panic(err)
	}

	userSelect := widget.NewSelect(availableUsers, nil)
	userSelect.Selected = availableUsers[0]

	passwordField := widget.NewPasswordEntry()

	form := widget.NewForm(
		widget.NewFormItem("User", userSelect),
		widget.NewFormItem("Password", passwordField),
	)
	form.SubmitText = "Login"

	submitFunc := func() {
		username := userSelect.Selected
		password := passwordField.Text

		credentials, err := pki.GetUserCredentials(username, []byte(password))
		if err != nil {
			if errors.Is(err, pki.ErrWrongPassword) {
				dialog.ShowError(fmt.Errorf("wrong password"), w)
				return
			} else {
				panic(err)
			}
		}

		startMainMenu(w, credentials)
	}

	form.OnSubmit = submitFunc
	passwordField.OnSubmitted = func(s string) {
		submitFunc()
	}

	w.SetContent(container.NewVBox(
		infoLabel,
		form,
	))

	w.Canvas().Focus(passwordField)
}

func setup(w fyne.Window) {
	addressInput := widget.NewEntry()
	addressInput.Text = "localhost:1234"

	form := widget.NewForm(
		widget.NewFormItem("Server Address", addressInput),
	)

	form.SubmitText = "Connect"

	form.OnSubmit = func() {

		address := addressInput.Text

		conn, err := rpc.FirstClientConnect(address)
		if err != nil {
			panic(err)
		}

		switch conn.GetProtocol() {
		case rpc.ProtoClientLogin:
			setupLoginForm(w, conn)

		case rpc.ProtoServerInit:
			setupServerForm(w, conn)

		default:
			conn.Close(400, "")
			panic(fmt.Errorf("unknown protocol %v", conn.GetProtocol()))
		}
	}

	w.SetContent(
		form,
	)
}

func setupLoginForm(w fyne.Window, conn *rpc.RpcConnection) {

	userInput := widget.NewEntry()

	passwordInput := widget.NewPasswordEntry()
	passwordInput.Validator = func(s string) error {
		if len(s) < 8 {
			return fmt.Errorf("password must be at least 8 characters")
		}

		return nil
	}

	totpInput := widget.NewEntry()
	totpInput.PlaceHolder = "00000000"
	totpInput.Validator = validation.NewRegexp("^[0-9]{8}$", "invalid TOTP code")

	form := widget.NewForm(
		widget.NewFormItem("User", userInput),
		widget.NewFormItem("Password", passwordInput),
		widget.NewFormItem("TOTP", totpInput),
	)

	form.OnSubmit = func() {

		username := userInput.Text

		password := passwordInput.Text

		totpCode := totpInput.Text

		credentials, err := rpc.Login(conn, username, []byte(password), totpCode)
		if err != nil {
			panic(err)
		}

		startMainMenu(w, credentials)
	}

	w.SetContent(
		form,
	)
}

func setupServerForm(w fyne.Window, conn *rpc.RpcConnection) {

	serverNameInput := widget.NewEntry()

	userInput := widget.NewEntry()

	passwordInput := widget.NewPasswordEntry()
	passwordInput.Validator = func(s string) error {
		if len(s) < 8 {
			return fmt.Errorf("password must be at least 8 characters")
		}

		return nil
	}

	passwordRepeatInput := widget.NewPasswordEntry()
	passwordRepeatInput.Validator = func(s string) error {
		password := passwordInput.Text

		if password != s {
			return fmt.Errorf("passwords do not match")
		}

		return nil
	}

	form := widget.NewForm(
		widget.NewFormItem("Server Name", serverNameInput),
		widget.NewFormItem("Root User", userInput),
		widget.NewFormItem("Password", passwordInput),
		widget.NewFormItem("Repeat Password", passwordRepeatInput),
	)

	form.OnSubmit = func() {
		go func() {
			totpCode, totpSecret, err := askForNewTotp(userInput.Text, w.Canvas())
			if err != nil {
				log.Printf("error generating totp: %v", err)
				return
			}

			fmt.Printf("\ntotp secret: %s\ntotp code: %s\n", totpSecret, totpCode)

			serverName := serverNameInput.Text

			username := userInput.Text

			password := passwordInput.Text

			err = pki.InitRoot(username, []byte(password))
			if err != nil {
				log.Printf("failed to init root: %v", err)
				return
			}

			credentials, err := pki.GetUserCredentials(username, []byte(password))
			if err != nil {
				log.Printf("failed to get user credentials: %v", err)
				return
			}

			err = rpc.SetupServer(conn, credentials, serverName)
			if err != nil {
				log.Printf("failed to setup server: %v", err)
			}

			reg, err := rpc.NewRegisterUserCmd(credentials, []byte(password), totpSecret, totpCode)
			if err != nil {
				panic(err)
			}

			cli, err := rpc.ConnectToUpstream(context.Background(), credentials)
			if err != nil {
				panic(err)
			}

			err = cli.SendSyncCommand(context.Background(), reg)
			if err != nil {
				panic(err)
			}

			startMainMenu(w, credentials)
		}()
	}

	w.SetContent(
		form,
	)
}

func askForNewTotp(accountName string, targetCanvas fyne.Canvas) (code string, secret string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "github.com/rahn-it/svalin",
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

	wg := sync.WaitGroup{}
	wg.Add(1)

	errorLabel := &widget.TextSegment{Style: widget.RichTextStyle{ColorName: "red"}, Text: ""}

	form.OnSubmit = func() {
		var err error
		code, err = codeBinding.Get()
		if err != nil {
			log.Printf("error getting code: %v", err)
			return
		}

		if util.ValidateTotp(secret, code) {
			wg.Done()
			return
		} else {
			errorLabel.Text = "Invalid code"
		}
	}

	popup := widget.NewModalPopUp(
		container.NewVBox(
			qrCode,
			secretEntry,
			widget.NewRichText(errorLabel),
			form,
		),
		targetCanvas,
	)
	popup.Resize(fyne.NewSize(500, 500))
	popup.Show()

	wg.Wait()

	popup.Hide()

	return
}
