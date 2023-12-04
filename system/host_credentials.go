package system

import (
	"errors"
	"fmt"

	"github.com/rahn-it/svalin/db"
	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/util"
)

var hostPasswordKey = []byte("host-password")
var hostCredentialsKey = []byte("host-credentials")

func LoadHostCredentials(scope db.Scope) (*pki.PermanentCredentials, error) {
	var creds *pki.PermanentCredentials
	err := scope.View(func(b db.Bucket) error {
		password := b.Get(hostPasswordKey)
		if password == nil {
			return errors.New("host password not found")
		}

		raw := b.Get(hostCredentialsKey)
		if raw == nil {
			return errors.New("host credentials not found")
		}

		credentials, err := pki.CredentialsFromPem(raw, password)
		if err != nil {
			return fmt.Errorf("failed to parse credentials: %w", err)
		}

		creds = credentials
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load host credentials: %w", err)
	}

	return creds, nil
}

func SaveHostCredentials(b db.Bucket, credentials *pki.PermanentCredentials) error {
	if credentials == nil {
		return errors.New("credentials cannot be nil")
	}

	password, err := util.GeneratePassword()
	if err != nil {
		return fmt.Errorf("failed to generate password: %w", err)
	}

	raw, err := credentials.PemEncode(password)
	if err != nil {
		return fmt.Errorf("failed to encode credentials: %w", err)
	}

	err = b.Put(hostPasswordKey, password)
	if err != nil {
		return fmt.Errorf("failed to save password: %w", err)
	}

	err = b.Put(hostCredentialsKey, raw)
	if err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	return nil
}
