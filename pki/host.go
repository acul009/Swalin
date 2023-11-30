package pki

import (
	"errors"
	"fmt"
	"github.com/rahn-it/svalin/config"
	"io/fs"
)

const (
	passwordFilePath = "password.pem"
	hostFilname      = "host"
)

var ErrNotInitialized = hostNotInitializedError{}

type hostNotInitializedError struct {
}

func (e hostNotInitializedError) Error() string {
	return "server not yet initialized"
}

func (e hostNotInitializedError) Is(target error) bool {
	_, ok := target.(hostNotInitializedError)
	return ok
}

func getHostPassword() ([]byte, error) {
	password, err := loadPasswordFromFile(config.GetFilePath(passwordFilePath))
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("failed to load password: %w", err)
		}

		password, err = generatePassword()
		if err != nil {
			return nil, fmt.Errorf("failed to generate password: %w", err)
		}
		err = savePasswordToFile(config.GetFilePath(passwordFilePath), password)
		if err != nil {
			return nil, fmt.Errorf("failed to save password: %w", err)
		}
	}

	return password, nil
}

func GetHostCredentials() (*PermanentCredentials, error) {
	password, err := getHostPassword()
	if err != nil {
		return nil, fmt.Errorf("failed to get host credentials: %w", err)
	}

	credentials := getCredentials(password, hostFilname)
	if !credentials.Available() {
		return nil, hostNotInitializedError{}
	}

	return credentials, nil
}

func (t *TempCredentials) UpgradeToHostCredentials(cert *Certificate) (*PermanentCredentials, error) {
	password, err := getHostPassword()
	if err != nil {
		return nil, fmt.Errorf("failed to get host credentials: %w", err)
	}

	permanent, err := t.toPermanentCredentials(password, cert, hostFilname)
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade credentials: %w", err)
	}

	return permanent, nil
}
