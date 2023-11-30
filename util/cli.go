package util

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/term"
)

var reader = bufio.NewReader(os.Stdin)

func AskForNewPassword(message string) ([]byte, error) {

	for {
		text1, err := AskForPassword(message)
		if err != nil {
			return nil, fmt.Errorf("error reading password: %w", err)
		}

		text2, err := AskForPassword("Repeat Password")
		if err != nil {
			return nil, fmt.Errorf("error reading password: %w", err)
		}

		passwordsMatch := true

		if len(text1) != len(text2) {
			passwordsMatch = false
		} else {

			for i := 0; i < len(text1); i++ {
				if text1[i] != text2[i] {
					passwordsMatch = false
				}
			}

		}

		if passwordsMatch {
			// convert CRLF to LF
			return text1, nil
		} else {
			fmt.Println("\nPasswords did not match, try again!")
		}
	}
}

func AskForNewTotp(accountName string) (secret string, currentCode string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "github.com/rahn-it/svalin",
		AccountName: accountName,
		SecretSize:  32,
		Rand:        rand.Reader,
		Period:      30,
		Digits:      otp.DigitsEight,
	})

	if err != nil {
		return "", "", err
	}

	url := key.URL()

	fmt.Printf("Generating Totp for %s...\nPlease save this secret in your authenicator app:\n%s\n\n", accountName, url)
	for {
		currentCode, err = AskForTotpCode(accountName)
		if err != nil {
			return "", "", fmt.Errorf("error reading code: %w", err)
		}

		if ValidateTotp(url, currentCode) {
			return url, currentCode, nil
		}

		fmt.Println("OTP invalid, try again!")
	}

}

func ValidateTotp(url string, code string) bool {
	key, err := otp.NewKeyFromURL(url)

	if err != nil {
		log.Printf("Error decoding TOTP: %s", err)
		return false
	}

	log.Printf("TOTP: %+v", key)

	ok, err := totp.ValidateCustom(code, key.Secret(), time.Now().UTC(), totp.ValidateOpts{
		Period:    uint(key.Period()),
		Skew:      0,
		Digits:    key.Digits(),
		Algorithm: key.Algorithm(),
	})

	if err != nil {
		log.Printf("Error validating TOTP: %s", err)
		return false
	}

	return ok
}

func AskForTotpCode(accountName string) (string, error) {
	return AskForString(fmt.Sprintf("Enter TOTP for user %s", accountName))
}

func AskForPassword(message string) ([]byte, error) {
	fmt.Print(message + ": ")
	text, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, fmt.Errorf("error reading password: %w", err)
	}

	fmt.Print("\n")

	return text, nil
}

// Asks for a string on the terminal which is not censored
func AskForString(message string) (string, error) {
	fmt.Print(message + ": ")
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading string: %w", err)
	}
	return text[:len(text)-1], nil
}
