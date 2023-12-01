package server

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"github.com/rahn-it/svalin/db"
	"github.com/rahn-it/svalin/pki"
	"go.etcd.io/bbolt"
)

type serverConfig struct {
	scope       db.Scope
	seed        []byte
	credentials *pki.PermanentCredentials
}

func openServerConfig(scope db.Scope) (*serverConfig, error) {
	sc := &serverConfig{
		scope: scope,
	}

	err := sc.initSeed()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize seed: %w", err)
	}

	err = sc.loadCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	return sc, nil
}

func (sc *serverConfig) initSeed() error {
	return sc.scope.View(func(b *bbolt.Bucket) error {
		seed := b.Get([]byte("seed"))
		if seed == nil {
			seed := make([]byte, 32)
			_, err := io.ReadFull(rand.Reader, seed)
			if err != nil {
				return fmt.Errorf("failed to generate seed: %w", err)
			}

			err = b.Put([]byte("seed"), seed)
			if err != nil {
				return fmt.Errorf("failed to save seed: %w", err)
			}
		}

		sc.seed = seed
		return nil
	})

}

func (sc *serverConfig) Seed() []byte {
	return sc.seed
}

func (sc *serverConfig) loadCredentials() error {
	var password []byte
	var raw []byte
	err := sc.scope.View(func(b *bbolt.Bucket) error {
		password = b.Get([]byte("password"))
		if password == nil {
			return errors.New("password not found")
		}

		raw = b.Get([]byte("credentials"))
		if raw == nil {
			return errors.New("credentials not found")
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	credentials, err := pki.CredentialsFromPem(raw, password)
	if err != nil {
		return fmt.Errorf("failed to parse credentials: %w", err)
	}

	sc.credentials = credentials
	return nil
}

func (sc *serverConfig) Credentials() *pki.PermanentCredentials {
	return sc.credentials
}
