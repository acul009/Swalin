package pki

import (
	"crypto/ecdsa"
	"crypto/x509"
	"errors"
	"fmt"
	"io/fs"
	"rahnit-rmm/config"
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

	if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to load CA certificate: %v", err)
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

const (
	caCertFilePath = "ca.crt"
	caKeyFilePath  = "ca.key"
)
