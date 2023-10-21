package pki

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"rahnit-rmm/util"
)

func SavePublicKeyToFile(filepath string, key *ecdsa.PublicKey) error {
	err := util.CreateParentDir(filepath)
	if err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	pubFile, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create public key file: %w", err)
	}

	keyBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %w", err)
	}

	defer pubFile.Close()
	err = pem.Encode(pubFile, &pem.Block{Type: "PUBLIC KEY", Bytes: keyBytes})
	if err != nil {
		return fmt.Errorf("failed to encode public key: %w", err)
	}

	return nil
}

func LoadPublicKeyFromFile(filepath string) (*ecdsa.PublicKey, error) {
	// Read the public key file
	pubPEM, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	// Decode the PEM-encoded public key
	block, _ := pem.Decode(pubPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode public key PEM")
	}

	// Parse the public key
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	// Cast the public key to the correct type
	pubTyped, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not of type *ecdsa.PublicKey")
	}

	return pubTyped, nil
}

func SaveCertToFile(filepath string, cert *x509.Certificate) error {
	err := util.CreateParentDir(filepath)
	if err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	certFile, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}

	defer certFile.Close()
	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	if err != nil {
		return fmt.Errorf("failed to encode certificate: %w", err)
	}
	return nil
}

func LoadCertFromFile(filepath string) (*x509.Certificate, error) {
	// Read the certificate file
	certPEM, err := os.ReadFile(filepath)

	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	// Decode the PEM-encoded certificate
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	// Parse the CA certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return cert, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

func SaveCertKeyToFile(filepath string, key *ecdsa.PrivateKey, password []byte) error {
	err := util.CreateParentDir(filepath)
	if err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	encoded, err := SerializePrivateKey(key, password)
	if err != nil {
		return fmt.Errorf("failed to serialize private key: %w", err)
	}

	os.WriteFile(filepath, encoded, 0600)

	return nil
}

func SerializePrivateKey(key *ecdsa.PrivateKey, password []byte) ([]byte, error) {

	caKeyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	encryptedBytes, err := util.EncryptDataWithPassword(password, caKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	return pem.EncodeToMemory(
		&pem.Block{Type: "EC PRIVATE KEY",
			Bytes:   encryptedBytes,
			Headers: map[string]string{"Proc-Type": "4,ENCRYPTED", "DEK-Info": "AES-CFB"},
		},
	), nil
}

func LoadCertKeyFromFile(filepath string, password []byte) (*ecdsa.PrivateKey, error) {
	keyPEM, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return DeserializePrivateKey(keyPEM, password)
}

func DeserializePrivateKey(serialized []byte, password []byte) (*ecdsa.PrivateKey, error) {

	// Decode the PEM-encoded CA private key
	block, _ := pem.Decode(serialized)
	if block == nil {
		return nil, fmt.Errorf("failed to decode CA private key PEM")
	}

	decryptedData, err := util.DecryptDataWithPassword(password, block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt CA private key: %w", err)
	}

	// Parse the CA private key
	key, err := x509.ParseECPrivateKey(decryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA private key: %w", err)
	}

	return key, nil
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
