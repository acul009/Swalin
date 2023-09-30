package pki

import (
	"crypto/ecdsa"
	"crypto/x509"
	"errors"
	"fmt"
	"io/fs"
	"rahnit-rmm/config"
)

// Custom go error to indicate that the CA certificate is missing
type missingCaCertError struct {
	cause error
}

func (e missingCaCertError) Error() string {
	return fmt.Errorf("CA certificate not found: %w", e.cause).Error()
}

func (e missingCaCertError) Unwrap() error {
	return e.cause
}

var ErrMissingCaCert = missingCaCertError{}

func (e missingCaCertError) Is(target error) bool {
	_, ok := target.(missingCaCertError)
	return ok
}

const (
	caCertFilePath = "ca.crt"
	caKeyFilePath  = "ca.key"
)

func GetCa(password []byte) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	caCert, err := GetCaCert()
	if err != nil {
		return caCert, nil, fmt.Errorf("failed to load CA: %v", err)
	}
	caKey, err := GetCaKey(password)
	if err != nil {
		return caCert, caKey, fmt.Errorf("failed to load CA: %v", err)
	}
	return caCert, caKey, nil
}

func SaveCaCert(caCert *x509.Certificate) error {
	_, err := GetCaCert()
	if err == nil {
		return fmt.Errorf("CA certificate already exists")
	}

	if !errors.Is(err, ErrMissingCaCert) {
		return fmt.Errorf("failed to load existing CA certificate: %v", err)
	}

	err = SaveCertToFile(config.GetFilePath(caCertFilePath), caCert.Raw)
	if err != nil {
		return fmt.Errorf("failed to save CA certificate: %v", err)
	}
	return nil
}

func GetCaCert() (*x509.Certificate, error) {
	caCert, err := LoadCertFromFile(config.GetFilePath(caCertFilePath))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return caCert, missingCaCertError{cause: err}
		}
		return caCert, err
	}
	return caCert, nil
}

func GetCaKey(password []byte) (*ecdsa.PrivateKey, error) {
	caKey, err := LoadCertKeyFromFile(config.GetFilePath(caKeyFilePath), password)
	if err != nil {
		return caKey, fmt.Errorf("failed to load CA certificate: %v", err)
	}
	return caKey, nil
}

func InitCa(password []byte) error {
	if _, err := GetCaCert(); err == nil {
		return fmt.Errorf("CA certificate already exists")
	} else if !errors.Is(err, ErrMissingCaCert) {
		return fmt.Errorf("failed to load existing CA certificate: %v", err)
	}

	caCert, caKey, err := generateRootCert()
	if err != nil {
		return fmt.Errorf("failed to generate CA certificate: %v", err)
	}

	err = SaveCertToFile(config.GetFilePath(caCertFilePath), caCert.Raw)
	if err != nil {
		return fmt.Errorf("failed to save CA certificate: %v", err)
	}

	err = SaveCertKeyToFile(config.GetFilePath(caKeyFilePath), caKey, password)
	if err != nil {
		return fmt.Errorf("failed to save CA certificate: %v", err)
	}

	return nil
}
