package server

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"github.com/rahn-it/svalin/db"
	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/system"
)

type serverConfig struct {
	scope       db.Scope
	seed        []byte
	credentials *pki.PermanentCredentials
	root        *pki.Certificate
}

func openServerConfig(scope db.Scope) (*serverConfig, error) {
	sc := &serverConfig{
		scope: scope,
	}

	err := sc.initSeed()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize seed: %w", err)
	}

	err = sc.loadRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to load root: %w", err)
	}

	err = sc.loadCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	return sc, nil
}

func (sc *serverConfig) initSeed() error {
	return sc.scope.View(func(b db.Bucket) error {
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

		sc.seed = make([]byte, len(seed))
		copy(sc.seed, seed)
		return nil
	})

}

func (sc *serverConfig) Seed() []byte {
	return sc.seed
}

func (sc *serverConfig) loadCredentials() error {
	creds, err := system.LoadHostCredentials(sc.scope)
	if err != nil {
		return fmt.Errorf("failed to load host credentials: %w", err)
	}

	sc.credentials = creds
	return nil
}

func (sc *serverConfig) Credentials() *pki.PermanentCredentials {
	return sc.credentials
}

func (sc *serverConfig) loadRoot() error {
	return sc.scope.View(func(b db.Bucket) error {
		raw := b.Get([]byte("root"))

		if raw == nil {
			return errors.New("root certificate not found")
		}
		root, err := pki.CertificateFromPem(raw)
		if err != nil {
			return fmt.Errorf("failed to load root certificate: %w", err)
		}

		sc.root = root
		return nil
	})
}

func (sc *serverConfig) Root() *pki.Certificate {
	return sc.root
}

func checkForServerConfig(scope db.Scope) (bool, error) {
	found := false
	err := scope.View(func(b db.Bucket) error {
		root := b.Get([]byte("root"))
		if root != nil {
			found = true
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("failed to check for server config: %w", err)
	}

	return found, nil
}

func initServerConfig(scope db.Scope, credentials *pki.PermanentCredentials, root *pki.Certificate) error {
	return scope.Update(func(b db.Bucket) error {
		err := b.Put([]byte("root"), root.PemEncode())
		if err != nil {
			return fmt.Errorf("failed to save root certificate: %w", err)
		}

		err = system.SaveHostCredentials(b, credentials)
		if err != nil {
			return fmt.Errorf("failed to save host credentials: %w", err)
		}

		return nil
	})
}
