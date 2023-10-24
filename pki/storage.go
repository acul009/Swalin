package pki

import (
	"encoding/pem"
	"fmt"
	"os"
	"rahnit-rmm/util"
)

func (pub *PublicKey) SaveToFile(filepath string) error {
	err := util.CreateParentDir(filepath)
	if err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	encoded, err := pub.PemEncode()
	if err != nil {
		return fmt.Errorf("failed to serialize public key: %w", err)
	}

	err = os.WriteFile(filepath, []byte(encoded), 0600)
	if err != nil {
		return fmt.Errorf("failed to write public key file: %w", err)
	}

	return nil
}

func LoadPublicKeyFromFile(filepath string) (*PublicKey, error) {
	// Read the public key file
	pubPEM, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	pubKey, err := PublicKeyFromPem(pubPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	return pubKey, nil
}

func (cert *Certificate) SaveToFile(filepath string) error {
	err := util.CreateParentDir(filepath)
	if err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	err = os.WriteFile(filepath, cert.PemEncode(), 0600)
	if err != nil {
		return fmt.Errorf("failed to write certificate file: %w", err)
	}

	return nil
}

func LoadCertificateFromFile(filepath string) (*Certificate, error) {
	// Read the certificate file
	certPEM, err := os.ReadFile(filepath)

	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	cert, err := CertificateFromPem(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to decode certificate: %w", err)
	}

	return cert, nil
}

func (key *PrivateKey) SaveToFile(filepath string, password []byte) error {
	err := util.CreateParentDir(filepath)
	if err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	encoded, err := key.PemEncode(password)
	if err != nil {
		return fmt.Errorf("failed to serialize private key: %w", err)
	}

	err = os.WriteFile(filepath, []byte(encoded), 0600)
	if err != nil {
		return fmt.Errorf("failed to write private key file: %w", err)
	}

	return nil
}

func LoadPrivateKeyFromFile(filepath string, password []byte) (*PrivateKey, error) {
	keyPEM, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return PrivateKeyFromPem(keyPEM, password)
}

func SavePasswordToFile(filepath string, password []byte) error {
	err := util.CreateParentDir(filepath)
	if err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	certFile, err := os.Create(filepath)
	if err != nil {
		return err
	}

	defer certFile.Close()
	err = pem.Encode(certFile, &pem.Block{Type: "Password", Bytes: password})
	if err != nil {
		return fmt.Errorf("failed to encode certificate: %w", err)
	}
	return nil
}

func LoadPasswordFromFile(filepath string) ([]byte, error) {
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

	return block.Bytes, nil
}
