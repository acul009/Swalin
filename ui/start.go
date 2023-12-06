package ui

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rahn-it/svalin/config"
	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/rpc"
	"github.com/rahn-it/svalin/system"
	"github.com/rahn-it/svalin/system/client"
	"github.com/rahn-it/svalin/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/cmd/fyne_demo/data"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func StartUI() {

	a := app.NewWithID("io.fyne.demo")
	a.SetIcon(data.FyneLogo)
	w := a.NewWindow("Fyne Demo")

	chooseProfile(w)

	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}

func chooseProfile(w fyne.Window) {
	profiles, err := config.ListProfiles("client")
	if err != nil {
		panic(err)
	}

	if len(profiles) == 0 {
		setup(w)
		return
	}

	profileSelect := widget.NewSelect(profiles, nil)

	selectButton := widget.NewButton("Select", func() {

	})
	selectButton.Disable()

	deleteButton := widget.NewButton("Delete", func() {
		dialog.NewConfirm(
			"Confirm Deletion",
			fmt.Sprintf("Are you sure you want to delete the following profile?\n%s", profileSelect.Selected),
			func(b bool) {
				if b {
					profile := profileSelect.Selected
					profileSelect.Selected = ""

					err := config.DeleteProfile(profile, "client")
					if err != nil {
						log.Printf("error deleting profile %s: %w", profile, err)
					}

					newProfiles := make([]string, 0, len(profiles)-1)

					for _, p := range profiles {
						if p != profile {
							newProfiles = append(newProfiles, p)
						}
					}

					profileSelect.SetOptions(newProfiles)
					profileSelect.OnChanged("")
					profileSelect.Refresh()
				}
			},
			w,
		).Show()
	})
	deleteButton.Disable()

	profileSelect.OnChanged = func(s string) {
		if s == "" {
			selectButton.Disable()
			deleteButton.Disable()
		} else {
			selectButton.Enable()
			deleteButton.Enable()
		}
	}

	newButton := widget.NewButtonWithIcon("New", theme.ContentAddIcon(), func() {
		setup(w)
	})

	w.SetContent(container.NewVBox(
		widget.NewLabel("Choose profile"),
		profileSelect,
		selectButton,
		deleteButton,
		layout.NewSpacer(),
		newButton,
	))
}

func unlock(w fyne.Window) {
	infoLabel := widget.NewLabel("Please login")

	availableUsers := []string{
		"admin",
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
			setupLoginForm(w, address, conn)

		case rpc.ProtoServerInit:
			setupServerForm(w, address, conn)

		default:
			conn.Close(400, "")
			panic(fmt.Errorf("unknown protocol %v", conn.GetProtocol()))
		}
	}

	w.SetContent(
		form,
	)
}

func setupLoginForm(w fyne.Window, addr string, conn *rpc.RpcConnection) {

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
		executor := system.NewLoginExecutor(username, []byte(password), totpCode, func(epii *rpc.EndPointInitInfo) error {

			profilename := fmt.Sprintf("%s@%s", username, addr)

			log.Printf("login request approved, setting up profile %s", profilename)

			profile, err := config.OpenProfile(profilename, "client")
			if err != nil {
				return fmt.Errorf("failed to open profile: %w", err)
			}

			log.Printf("profile %s opened", profilename)

			err = client.SetupClient(profile, epii.Root, epii.Upstream, epii.Credentials, []byte(password), addr)
			if err != nil {
				return fmt.Errorf("failed to setup client profile: %w", err)
			}

			log.Printf("client setup complete for profile %s", profile.Name())

			return nil
		})

		go func() {
			err := rpc.Login(conn, executor.Login)
			if err != nil {
				panic(err)
			}

			log.Printf("login finished, continuing...")
		}()
	}

	w.SetContent(
		form,
	)
}

func setupServerForm(w fyne.Window, addr string, conn *rpc.RpcConnection) {

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

			credentials, err := pki.GenerateRootCredentials(username)
			if err != nil {
				log.Printf("failed to generate root credentials: %v", err)
				return
			}

			upstream, err := rpc.SetupServer(conn, credentials, serverName)
			if err != nil {
				log.Printf("failed to setup server: %v", err)
				return
			}

			time.Sleep(time.Second)

			regCmd, err := system.NewRegisterUserCmd(credentials, []byte(password), totpSecret, totpCode)
			if err != nil {
				log.Printf("failed to create register user command: %v", err)
				return
			}

			verifier, err := system.NewFallbackVerifier(credentials.Certificate(), upstream)
			if err != nil {
				log.Printf("failed to create fallback verifier: %v", err)
				return
			}

			ep, err := rpc.ConnectToServer(context.Background(), addr, credentials, upstream, verifier)
			if err != nil {
				log.Printf("failed to connect to server: %v", err)
				return
			}

			err = ep.SendSyncCommand(context.Background(), regCmd)
			if err != nil {
				log.Printf("failed to register user: %v", err)
				return
			}

			profilename := fmt.Sprintf("%s@%s", username, addr)

			profile, err := config.OpenProfile(profilename, "client")
			if err != nil {
				log.Printf("failed to open profile: %v", err)
				return
			}

			err = client.SetupClient(profile, credentials.Certificate(), upstream, credentials, []byte(password), addr)
			if err != nil {
				log.Printf("failed to setup client profile: %v", err)
				return
			}

			log.Printf("Successfully registered user %s on server %s", username, serverName)

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
