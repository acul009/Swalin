package pki

import (
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

func GetRoot(password []byte) (*Certificate, *PrivateKey, error) {
	rootCert, err := GetRootCert()
	if err != nil {
		return rootCert, nil, fmt.Errorf("failed to load CA: %w", err)
	}
	rootKey, err := GetRootKey(password)
	if err != nil {
		return rootCert, rootKey, fmt.Errorf("failed to load CA: %w", err)
	}
	return rootCert, rootKey, nil
}

func SaveRootCert(rootCert *Certificate) error {
	_, err := GetRootCert()
	if err == nil {
		return fmt.Errorf("root certificate already exists")
	}

	if !errors.Is(err, ErrNoRootCert) {
		return fmt.Errorf("failed to load existing root certificate: %w", err)
	}

	err = rootCert.SaveToFile(config.GetFilePath(rootCertFilePath))
	if err != nil {
		return fmt.Errorf("failed to save root certificate: %w", err)
	}

	updateRootPool()

	return nil
}

func GetRootCert() (*Certificate, error) {
	rootCert, err := LoadCertificateFromFile(config.GetFilePath(rootCertFilePath))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return rootCert, noRootCertError{cause: err}
		}
		return rootCert, err
	}
	return rootCert, nil
}

func GetRootKey(password []byte) (*PrivateKey, error) {
	rootKey, err := LoadPrivateKeyFromFile(config.GetFilePath(rootKeyFilePath), password)
	if err != nil {
		return rootKey, fmt.Errorf("failed to load CA certificate: %w", err)
	}
	return rootKey, nil
}

func IsRootPublicKey(pub *PublicKey) (bool, error) {
	caCert, err := GetRootCert()
	if err != nil {
		return false, fmt.Errorf("failed to load CA certificate: %w", err)
	}

	return pub.Equal(caCert.GetPublicKey()), nil
}

func InitRoot(rootName string, password []byte) error {
	if _, err := GetRootCert(); err == nil {
		return fmt.Errorf("root certificate already exists")
	} else if !errors.Is(err, ErrNoRootCert) {
		return fmt.Errorf("failed to load existing root certificate: %w", err)
	}

	rootCert, rootKey, err := generateRootCert(rootName)
	if err != nil {
		return fmt.Errorf("failed to generate root certificate: %w", err)
	}

	err = rootCert.SaveToFile(config.GetFilePath(rootCertFilePath))
	if err != nil {
		return fmt.Errorf("failed to save root certificate: %w", err)
	}

	err = rootKey.SaveToFile(config.GetFilePath(rootKeyFilePath), password)
	if err != nil {
		return fmt.Errorf("failed to save root certificate: %w", err)
	}

	return nil
}
