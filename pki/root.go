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
type noRootCertError struct {
	cause error
}

func (e noRootCertError) Error() string {
	return fmt.Errorf("root certificate not found: %w", e.cause).Error()
}

func (e noRootCertError) Unwrap() error {
	return e.cause
}

var ErrNoRootCert = noRootCertError{}

func (e noRootCertError) Is(target error) bool {
	_, ok := target.(noRootCertError)
	return ok
}

const (
	rootCertFilePath = "root.crt"
	rootKeyFilePath  = "root.key"
)

func GetRoot(password []byte) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	caCert, err := GetRootCert()
	if err != nil {
		return caCert, nil, fmt.Errorf("failed to load CA: %w", err)
	}
	caKey, err := GetRootKey(password)
	if err != nil {
		return caCert, caKey, fmt.Errorf("failed to load CA: %w", err)
	}
	return caCert, caKey, nil
}

func SaveRootCert(caCert *x509.Certificate) error {
	_, err := GetRootCert()
	if err == nil {
		return fmt.Errorf("root certificate already exists")
	}

	if !errors.Is(err, ErrNoRootCert) {
		return fmt.Errorf("failed to load existing root certificate: %w", err)
	}

	err = SaveCertToFile(config.GetFilePath(rootCertFilePath), caCert)
	if err != nil {
		return fmt.Errorf("failed to save root certificate: %w", err)
	}

	updateRootPool()

	return nil
}

func GetRootCert() (*x509.Certificate, error) {
	caCert, err := LoadCertFromFile(config.GetFilePath(rootCertFilePath))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return caCert, noRootCertError{cause: err}
		}
		return caCert, err
	}
	return caCert, nil
}

func GetRootKey(password []byte) (*ecdsa.PrivateKey, error) {
	caKey, err := LoadCertKeyFromFile(config.GetFilePath(rootKeyFilePath), password)
	if err != nil {
		return caKey, fmt.Errorf("failed to load CA certificate: %w", err)
	}
	return caKey, nil
}

func IsRootPublicKey(pub *ecdsa.PublicKey) (bool, error) {
	caCert, err := GetRootCert()
	if err != nil {
		return false, fmt.Errorf("failed to load CA certificate: %w", err)
	}

	return pub.Equal(caCert.PublicKey), nil
}

func InitRoot(rootName string, password []byte) error {
	if _, err := GetRootCert(); err == nil {
		return fmt.Errorf("root certificate already exists")
	} else if !errors.Is(err, ErrNoRootCert) {
		return fmt.Errorf("failed to load existing root certificate: %w", err)
	}

	caCert, caKey, err := generateRootCert(rootName)
	if err != nil {
		return fmt.Errorf("failed to generate root certificate: %w", err)
	}

	err = SaveCertToFile(config.GetFilePath(rootCertFilePath), caCert)
	if err != nil {
		return fmt.Errorf("failed to save root certificate: %w", err)
	}

	err = SaveCertKeyToFile(config.GetFilePath(rootKeyFilePath), caKey, password)
	if err != nil {
		return fmt.Errorf("failed to save root certificate: %w", err)
	}

	return nil
}
