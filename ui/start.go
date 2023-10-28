package ui

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"rahnit-rmm/config"
	"rahnit-rmm/pki"
	"rahnit-rmm/rpc"
	"rahnit-rmm/util"
	"regexp"
	"sync"

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
	err := config.SetSubdir("client")
	if err != nil {
		panic(err)
	}

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

	w.Resize(fyne.NewSize(800, 600))
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
	addressBind.Set("localhost:1234")
	addressInput := widget.NewEntryWithData(addressBind)

	form := widget.NewForm(
		widget.NewFormItem("Server Address", addressInput),
	)

	form.SubmitText = "Connect"

	form.OnSubmit = func() {

		address, err := addressBind.Get()
		if err != nil {
			log.Printf("failed to get address: %v", err)
			return
		}

		conn, err := rpc.FirstClientConnect(address)
		if err != nil {
			log.Printf("failed to connect: %v", err)
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
			log.Printf("failed to get username: %v", err)
			return
		}

		password, err := passwordBind.Get()
		if err != nil {
			log.Printf("failed to get password: %v", err)
			return
		}

		totpCode, err := totpCodeBind.Get()
		if err != nil {
			log.Printf("failed to get TOTP code: %v", err)
			return
		}

		err = rpc.Login(conn, username, []byte(password), totpCode)
		if err != nil {
			log.Printf("login failed: %v", err)
		}
	}

	master.SetContent(
		form,
	)
}

func setupServerForm(conn *rpc.RpcConnection) {

	serverNameBind := binding.NewString()
	serverNameInput := widget.NewEntryWithData(serverNameBind)

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
		widget.NewFormItem("Server Name", serverNameInput),
		widget.NewFormItem("Root User", userInput),
		widget.NewFormItem("Password", passwordInput),
		widget.NewFormItem("Repeat Password", passwordRepeatInput),
	)

	form.OnSubmit = func() {
		go func() {
			totpCode, totpSecret, err := askForNewTotp(userInput.Text, master.Canvas())
			if err != nil {
				log.Printf("error generating totp: %v", err)
				return
			}

			fmt.Printf("\ntotp secret: %s\ntotp code: %s\n", totpSecret, totpCode)

			serverName, err := serverNameBind.Get()
			if err != nil {
				log.Printf("failed to get server name: %v", err)
				return
			}

			username, err := userBind.Get()
			if err != nil {
				log.Printf("failed to get username: %v", err)
				return
			}

			password, err := passwordBind.Get()
			if err != nil {
				log.Printf("failed to get password: %v", err)
				return
			}

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

			ep, err := rpc.ConnectToUpstream(context.Background(), credentials)
			if err != nil {
				panic(err)
			}

			session, err := ep.Session(context.Background())
			if err != nil {
				panic(err)
			}

			err = session.SendCommand(reg)
			if err != nil {
				panic(err)
			}
		}()
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
		master.Canvas(),
	)
	popup.Resize(fyne.NewSize(500, 500))
	popup.Show()

	wg.Wait()

	popup.Hide()

	return
}
