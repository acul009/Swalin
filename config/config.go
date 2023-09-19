package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"rahnit-rmm/util"
	"time"
)

func getConfigDir() string {
	if os.Getenv("OS") == "Windows_NT" {
		return filepath.Join(os.Getenv("APPDATA"), "rahnit-rmm")
	}
	return "/etc/rahnit-rmm"
}

func getConfigFilePath(filePath ...string) string {
	pathParts := make([]string, 1, len(filePath)+1)
	pathParts[0] = getConfigDir()
	pathParts = append(pathParts, filePath...)
	fullPath := filepath.Join(pathParts...)
	return fullPath
}

func GetCaCert() (*x509.Certificate, error) {
	// Read the CA certificate file
	caCertPEM, err := os.ReadFile(getConfigFilePath("ca.crt"))

	if err != nil {
		return nil, err
	}

	// Decode the PEM-encoded CA certificate
	block, _ := pem.Decode(caCertPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode CA certificate PEM")
	}

	// Parse the CA certificate
	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	return caCert, nil
}

const (
	caCertFilePath = "ca.crt"
	caKeyFilePath  = "ca.key"
	validFor       = 10 * 365 * 24 * time.Hour
)

func GenerateRootCert(password []byte) error {
	// check if the CA certificate already exists
	_, err := GetCaCert()
	if err == nil {
		return fmt.Errorf("CA certificate already exists")
	}

	// Generate a new CA private key
	caPrivateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return err
	}

	// Create a self-signed CA certificate template
	caTemplate := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Root CA"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(validFor),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Create the config dir
	os.MkdirAll(getConfigDir(), os.ModePerm)

	// Create and save the self-signed CA certificate
	caCertDER, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return err
	}

	caCertFile, err := os.Create(getConfigFilePath(caCertFilePath))
	if err != nil {
		return err
	}

	defer caCertFile.Close()
	err = pem.Encode(caCertFile, &pem.Block{Type: "CERTIFICATE", Bytes: caCertDER})
	if err != nil {
		return err
	}

	// Save the CA private key to a file
	caKeyFile, err := os.Create(getConfigFilePath(caKeyFilePath))
	if err != nil {
		return err
	}
	defer caKeyFile.Close()

	caKeyBytes, err := x509.MarshalECPrivateKey(caPrivateKey)
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

	fmt.Printf("CA certificate and private key saved to %v and %v\n", caCertFilePath, caKeyFilePath)

	return nil
}
