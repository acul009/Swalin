package util

import (
	"bufio"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/term"
)

func AskForNewPassword(message string) ([]byte, error) {

	for {
		fmt.Println(message)
		text1, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return nil, fmt.Errorf("error reading password: %v", err)
		}
		fmt.Println("Repeat Password:")
		text2, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return nil, fmt.Errorf("error reading password: %v", err)
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

func AskForPassword(message string) ([]byte, error) {
	fmt.Println(message)
	text, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, fmt.Errorf("error reading password: %v", err)
	}

	return text, nil
}

var reader = bufio.NewReader(os.Stdin)

// Asks for a string on the terminal which is not censored
func AskForString(message string) (string, error) {
	fmt.Println(message)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading string: %v", err)
	}
	return text[:len(text)-1], nil
}
