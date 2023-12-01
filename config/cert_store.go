package config

import (
	"errors"
	"fmt"

	"github.com/rahn-it/svalin/db"
	"github.com/rahn-it/svalin/pki"
	"go.etcd.io/bbolt"
)

type CertStore struct {
	scope           db.Scope
	credentialScope db.Scope
	root            *pki.Certificate
}

func openCertStore(scope db.Scope) (*CertStore, error) {

	cs := &CertStore{
		scope:           scope,
		credentialScope: scope.Scope([]byte("credentials")),
	}

	root, err := cs.LoadCert([]byte("root"))
	if err != nil {
		return nil, fmt.Errorf("failed to load root certificate: %w", err)
	}

	cs.root = root

	return cs, nil
}

func initCertStore(scope db.Scope, username string, password []byte) (*CertStore, error) {
	cs := &CertStore{
		scope:           scope,
		credentialScope: scope.Scope([]byte("credentials")),
	}

	found := false
	err := scope.View(func(b *bbolt.Bucket) error {
		if b.Get([]byte("root")) != nil {
			found = true
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if found {
		return nil, errors.New("root certificate already exists")
	}

	rootCredentials, err := pki.GenerateRootCredentials(username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate root certificate: %w", err)
	}

	err = cs.SaveCredentials(rootCredentials, password)
	if err != nil {
		return nil, fmt.Errorf("failed to save root credentials: %w", err)
	}

	err = cs.SaveCert([]byte("root"), rootCredentials.Certificate())
	if err != nil {
		return nil, fmt.Errorf("failed to save root certificate: %w", err)
	}

	cs.root = rootCredentials.Certificate()

	return cs, nil
}

func (cs *CertStore) LoadCredentials(name string, password []byte) (*pki.PermanentCredentials, error) {
	var raw []byte

	err := cs.credentialScope.View(func(b *bbolt.Bucket) error {
		raw = b.Get([]byte(name))
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	credentials, err := pki.CredentialsFromPem(raw, password)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return credentials, nil
}

func (cs *CertStore) SaveCredentials(credentials *pki.PermanentCredentials, password []byte) error {

	raw, err := credentials.PemEncode(password)
	if err != nil {
		return fmt.Errorf("failed to encode credentials: %w", err)
	}

	name := credentials.GetName()

	err = cs.credentialScope.Update(func(b *bbolt.Bucket) error {
		return b.Put([]byte(name), raw)
	})
	if err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	return nil
}

func (cs *CertStore) SaveCert(key []byte, cert *pki.Certificate) error {
	raw := cert.PemEncode()
	err := cs.scope.Update(func(b *bbolt.Bucket) error {
		return b.Put(key, raw)
	})

	if err != nil {
		return fmt.Errorf("failed to save certificate: %w", err)
	}

	return nil
}

func (cs *CertStore) LoadCert(key []byte) (*pki.Certificate, error) {
	var raw []byte
	err := cs.scope.View(func(b *bbolt.Bucket) error {
		raw = b.Get(key)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	if raw == nil {
		return nil, errors.New("certificate not found")
	}

	cert, err := pki.CertificateFromPem(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

func (cs *CertStore) Root() *pki.Certificate {
	return cs.root
}
