package pki

import (
	"crypto/ecdsa"
	"crypto/x509"
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
			return fmt.Errorf("failed to load password: %w", err)
		}

		password, err = generatePassword()
		if err != nil {
			return fmt.Errorf("failed to generate password: %w", err)
		}
		err = SavePasswordToFile(config.GetFilePath(passwordFilePath), password)
		if err != nil {
			return fmt.Errorf("failed to save password: %w", err)
		}
	}

	ready, err := CurrentAvailable()
	if err != nil {
		return fmt.Errorf("failed to check if current cert exists: %w", err)
	}

	if !ready {
		key, err := generateKeypair()
		if err != nil {
			return fmt.Errorf("failed to generate new keypair: %w", err)
		}
		SaveCurrentKeyPair(key, &key.PublicKey, password)
	}

	err = Unlock(password)
	if err != nil {
		return fmt.Errorf("failed to unlock: %w", err)
	}

	return nil
}

func CreateServerCertWithCurrent(name string, pub *ecdsa.PublicKey) (*x509.Certificate, error) {
	cert, err := GetCurrentCert()
	if err != nil {
		return nil, fmt.Errorf("failed to load current cert: %w", err)
	}

	key, err := GetCurrentKey()
	if err != nil {
		return nil, fmt.Errorf("failed to load current key: %w", err)
	}

	if !cert.IsCA {
		return nil, fmt.Errorf("current cert is not a CA")
	}

	serverCert, err := createServerCert(name, pub, key, cert)
	if err != nil {
		return nil, fmt.Errorf("failed to create server cert: %w", err)
	}

	return serverCert, nil
}
