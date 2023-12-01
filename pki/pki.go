package pki

import (
	"errors"
	"fmt"
)

type PKI struct {
	storage PkiStorage
	root    *Certificate
}

func Open(storage PkiStorage) (*PKI, error) {

	root, err := storage.LoadCertificate([]byte("root"))
	if err != nil {
		if errors.Is(err, &notFoundError{}) {
			root = nil
		} else {
			return nil, fmt.Errorf("failed to load root certificate: %w", err)
		}
	}

	return &PKI{
		storage: storage,
		root:    root,
	}, nil
}

type PkiStorage interface {
	LoadCertificate(key []byte) (*Certificate, error)
	SaveCertificate(key []byte, cert *Certificate) error
	LoadCredentials(key []byte, password []byte) (*PermanentCredentials, error)
	SaveCredentials(key []byte, credentials *PermanentCredentials) error
}

var ErrNotFound = &notFoundError{}

type notFoundError struct {
}

func (e *notFoundError) Error() string {
	return "not found"
}

func (e *notFoundError) Is(target error) bool {
	_, ok := target.(*notFoundError)
	return ok
}
