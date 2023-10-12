package pki

import (
	"errors"
	"fmt"
	"io/fs"
	"rahnit-rmm/config"
)

const (
	passwordFilePath = "password.pem"
)

// UnlockHost tries to unlock the machine certificate.
// since this needs to happen autonomously, it uses a password loaded from a file
// if there is no keypair or certificate, it creates a newpair by itself.
func UnlockHost() error {
	password, err := LoadPasswordFromFile(config.GetFilePath(passwordFilePath))
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("failed to load password: %v", err)
		}

		password, err = generatePassword()
		if err != nil {
			return fmt.Errorf("failed to generate password: %v", err)
		}
		err = SavePasswordToFile(config.GetFilePath(passwordFilePath), password)
		if err != nil {
			return fmt.Errorf("failed to save password: %v", err)
		}
	}

	ready, err := CurrentAvailable()
	if err != nil {
		return fmt.Errorf("failed to check if current cert exists: %v", err)
	}

	if !ready {
		key, err := generateKeypair()
		if err != nil {
			return fmt.Errorf("failed to generate new keypair: %v", err)
		}
		SaveCurrentKeyPair(key, &key.PublicKey, password)
	}

	err = Unlock(password)
	if err != nil {
		return fmt.Errorf("failed to unlock: %v", err)
	}

	return nil
}
