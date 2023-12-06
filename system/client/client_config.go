package client

import (
	"errors"
	"fmt"

	"github.com/rahn-it/svalin/db"
	"github.com/rahn-it/svalin/pki"
)

type clientConfig struct {
	scope       db.Scope
	root        *pki.Certificate
	upstream    *pki.Certificate
	credentials *pki.PermanentCredentials
	serverAddr  string
}

func openClientConfig(scope db.Scope, password []byte) (*clientConfig, error) {
	conf := &clientConfig{
		scope: scope,
	}

	err := conf.loadRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to load root: %w", err)
	}

	err = conf.loadUpstream()
	if err != nil {
		return nil, fmt.Errorf("failed to load upstream: %w", err)
	}

	err = conf.loadCredentials(password)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	err = conf.loadSererAddr()
	if err != nil {
		return nil, fmt.Errorf("failed to load server address: %w", err)
	}

	return conf, nil
}

func (conf *clientConfig) loadRoot() error {
	return conf.scope.View(func(b db.Bucket) error {
		raw := b.Get([]byte("root"))
		if raw == nil {
			return errors.New("root certificate not found")
		}
		root, err := pki.CertificateFromPem(raw)
		if err != nil {
			return fmt.Errorf("failed to load root certificate: %w", err)
		}

		conf.root = root
		return nil
	})
}

func (conf *clientConfig) Root() *pki.Certificate {
	return conf.root
}

func (conf *clientConfig) loadUpstream() error {
	return conf.scope.View(func(b db.Bucket) error {
		raw := b.Get([]byte("upstream"))
		if raw == nil {
			return errors.New("upstream certificate not found")
		}
		upstream, err := pki.CertificateFromPem(raw)
		if err != nil {
			return fmt.Errorf("failed to load upstream certificate: %w", err)
		}

		conf.upstream = upstream
		return nil
	})
}

func (conf *clientConfig) Upstream() *pki.Certificate {
	return conf.upstream
}

func (conf *clientConfig) loadCredentials(password []byte) error {
	return conf.scope.View(func(b db.Bucket) error {
		raw := b.Get([]byte("credentials"))
		if raw == nil {
			return errors.New("credentials not found")
		}
		credentials, err := pki.CredentialsFromPem(raw, password)
		if err != nil {
			return fmt.Errorf("failed to load credentials: %w", err)
		}

		conf.credentials = credentials
		return nil
	})
}

func (conf *clientConfig) Credentials() *pki.PermanentCredentials {
	return conf.credentials
}

func (conf *clientConfig) loadSererAddr() error {
	return conf.scope.View(func(b db.Bucket) error {
		raw := b.Get([]byte("serverAddr"))
		if raw == nil {
			return errors.New("server address not found")
		}

		conf.serverAddr = string(raw)
		return nil
	})
}

func (conf *clientConfig) ServerAddr() string {
	return conf.serverAddr
}

func initClientConfig(scope db.Scope, root *pki.Certificate, upstream *pki.Certificate, credentials *pki.PermanentCredentials, password []byte, serverAddr string) error {
	return scope.Update(func(b db.Bucket) error {
		currentRoot := b.Get([]byte("root"))
		if currentRoot != nil {
			return errors.New("config already initialized")
		}

		err := b.Put([]byte("root"), root.PemEncode())
		if err != nil {
			return fmt.Errorf("failed to save root certificate: %w", err)
		}

		err = b.Put([]byte("upstream"), upstream.PemEncode())
		if err != nil {
			return fmt.Errorf("failed to save upstream certificate: %w", err)
		}

		rawCredentials, err := credentials.PemEncode(password)
		if err != nil {
			return fmt.Errorf("failed to encode credentials: %w", err)
		}

		err = b.Put([]byte("credentials"), rawCredentials)
		if err != nil {
			return fmt.Errorf("failed to save credentials: %w", err)
		}

		err = b.Put([]byte("serverAddr"), []byte(serverAddr))
		if err != nil {
			return fmt.Errorf("failed to save server address: %w", err)
		}

		return nil
	})
}
