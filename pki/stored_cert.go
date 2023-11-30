package pki

import (
	"errors"
	"fmt"
	"github.com/rahn-it/svalin/config"
	"io/fs"
	"log"
	"os"
)

type storedCertificate struct {
	certificate   *Certificate
	allowOverride bool
	filename      string
}

// Custom go error to indicate that the CA certificate is missing
type certMissingError struct {
	cause error
}

func (e certMissingError) Error() string {
	return fmt.Errorf("Certificate not found: %w", e.cause).Error()
}

func (e certMissingError) Unwrap() error {
	return e.cause
}

var ErrCertMissing = certMissingError{}

func (s *storedCertificate) Available() bool {
	_, err := os.Stat(s.path())
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		log.Printf("failed to check if certificate exists: %v", err)
		return false
	}
	return true
}

func (s *storedCertificate) path() string {
	return config.GetFilePath(s.filename)
}

func (s *storedCertificate) Get() (*Certificate, error) {
	if s.certificate == nil {
		cert, err := loadCertificateFromFile(s.path())
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil, &certMissingError{
					cause: err,
				}
			}
			return nil, fmt.Errorf("failed to load certificate: %w", err)
		}

		s.certificate = cert
	}

	return s.certificate, nil
}

func (s *storedCertificate) Set(cert *Certificate) error {
	if !s.allowOverride && s.Available() {
		return errors.New("cannot override certificate")
	}

	err := cert.saveToFile(s.path())
	if err != nil {
		return fmt.Errorf("failed to save certificate: %w", err)
	}

	s.certificate = cert

	return nil
}

func (s *storedCertificate) GetPublicKey() (*PublicKey, error) {
	cert, err := s.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}

	return cert.GetPublicKey(), nil
}

func (s *storedCertificate) MatchesKey(pub *PublicKey) (bool, error) {
	myKey, err := s.GetPublicKey()
	if err != nil {
		return false, fmt.Errorf("failed to get public key: %w", err)
	}

	if !myKey.Equal(pub) {
		return false, nil
	}

	return true, nil
}
