package pki

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"rahnit-rmm/util"
)

func SaveCertToFile(filepath string, cert *x509.Certificate) error {
	err := util.CreateParentDir(filepath)
	if err != nil {
		return fmt.Errorf("failed to create parent directory: %v", err)
	}

	certFile, err := os.Create(filepath)
	if err != nil {
		return err
	}

	defer certFile.Close()
	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	if err != nil {
		return fmt.Errorf("failed to encode certificate: %v", err)
	}
	return nil
}

func LoadCertFromFile(filepath string) (*x509.Certificate, error) {
	// Read the certificate file
	certPEM, err := os.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	// Decode the PEM-encoded certificate
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	// Parse the CA certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return cert, err
	}

	return cert, nil
}

func SaveCertKeyToFile(filepath string, key *ecdsa.PrivateKey, password []byte) error {
	err := util.CreateParentDir(filepath)
	if err != nil {
		return fmt.Errorf("failed to create parent directory: %v", err)
	}

	caKeyFile, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create CA private key file: %v", err)
	}
	defer caKeyFile.Close()

	caKeyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return fmt.Errorf("failed to marshal CA private key: %v", err)
	}

	encryptedBytes, err := util.EncryptDataWithPassword(password, caKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to encrypt CA private key: %v", err)
	}

	err = pem.Encode(
		caKeyFile,
		&pem.Block{Type: "EC PRIVATE KEY",
			Bytes:   encryptedBytes,
			Headers: map[string]string{"Proc-Type": "4,ENCRYPTED", "DEK-Info": "AES-CFB"},
		})

	if err != nil {
		return err
	}

	return nil
}

func LoadCertKeyFromFile(filepath string, password []byte) (*ecdsa.PrivateKey, error) {
	caKeyPEM, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	// Decode the PEM-encoded CA private key
	block, _ := pem.Decode(caKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode CA private key PEM")
	}

	decryptedData, err := util.DecryptDataWithPassword(password, block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt CA private key: %v", err)
	}

	// Parse the CA private key
	caKey, err := x509.ParseECPrivateKey(decryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA private key: %v", err)
	}

	return caKey, nil
}
