package pki

import (
	"errors"
	"fmt"
	"io/fs"
	"rahnit-rmm/config"
)

const (
	passwordFilePath = "password.pem"
	hostFilname      = "host"
)

var ErrNotInitialized = serverNotInitializedError{}

type serverNotInitializedError struct {
}

func (e serverNotInitializedError) Error() string {
	return "server not yet initialized"
}

func (e serverNotInitializedError) Is(target error) bool {
	_, ok := target.(serverNotInitializedError)
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
		return nil, serverNotInitializedError{}
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
