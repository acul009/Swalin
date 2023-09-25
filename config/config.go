package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"io/fs"
	"math/big"
	"os"
	"path/filepath"
	"rahnit-rmm/util"
	"time"
)

var subdir = "default"

func SetSubdir(s string) {
	subdir = s
}

func GetSubdir() string {
	return subdir
}

func getConfigDir() string {
	if os.Getenv("OS") == "Windows_NT" {
		return filepath.Join(os.Getenv("APPDATA"), "rahnit-rmm", GetSubdir())
	}
	return filepath.Join("/etc/rahnit-rmm", GetSubdir())
}

func getConfigFilePath(filePath ...string) string {
	pathParts := make([]string, 1, len(filePath)+1)
	pathParts[0] = getConfigDir()
	pathParts = append(pathParts, filePath...)
	fullPath := filepath.Join(pathParts...)
	return fullPath
}

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

	err = util.SaveCert(getConfigFilePath("ca.crt"), caCert.Raw)
	if err != nil {
		return fmt.Errorf("failed to save CA certificate: %v", err)
	}
	return nil
}

func GetCaCert() (*x509.Certificate, error) {
	caCert, err := util.LoadCert(getConfigFilePath("ca.crt"))
	if err != nil {
		return caCert, err
	}
	return caCert, nil
}

func GetCaKey(password []byte) (*ecdsa.PrivateKey, error) {
	caKey, err := util.LoadCertKey(getConfigFilePath("ca.key"), password)
	if err != nil {
		return caKey, fmt.Errorf("failed to load CA certificate: %v", err)
	}
	return caKey, nil
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
		return fmt.Errorf("failed to generate CA private key: %v", err)
	}

	// Create a self-signed CA certificate template
	caTemplate := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Root CA"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(validFor),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Create and save the self-signed CA certificate
	caCertDER, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to create self-signed CA certificate: %v", err)
	}

	err = util.SaveCert(getConfigFilePath(caCertFilePath), caCertDER)
	if err != nil {
		return fmt.Errorf("failed to save CA certificate: %v", err)
	}

	// Save the CA private key to a file
	err = util.SaveCertKey(getConfigFilePath(caKeyFilePath), caPrivateKey, password)
	if err != nil {
		return fmt.Errorf("failed to save CA private key: %v", err)
	}

	fmt.Printf("CA certificate and private key saved to %v and %v\n", getConfigFilePath(caCertFilePath), getConfigFilePath(caKeyFilePath))

	return nil
}

func GenerateUserCert(username string, password []byte, rootPassword []byte) error {

	// Generate a new CA private key
	intermediateCAPrivateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return err
	}

	// Create a self-signed CA certificate template
	intermediateCATemplate := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: username},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(validFor),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caCert, caKey, err := GetCa(rootPassword)
	if err != nil {
		return fmt.Errorf("failed to load CA: %v", err)
	}

	// Create and sign the intermediate CA certificate
	userCACertificateDER, err := x509.CreateCertificate(rand.Reader, &intermediateCATemplate, caCert, &intermediateCAPrivateKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create user certificate: %v", err)
	}

	err = util.SaveCert(getConfigFilePath(caCertFilePath), userCACertificateDER)
	if err != nil {
		return fmt.Errorf("failed to save user certificate: %v", err)
	}

	// Save the CA private key to a file
	err = util.SaveCertKey(getConfigFilePath(caKeyFilePath), intermediateCAPrivateKey, password)
	if err != nil {
		return fmt.Errorf("failed to save user private key: %v", err)
	}

	fmt.Printf("user certificate and private key saved to %v and %v\n", getConfigFilePath(caCertFilePath), getConfigFilePath(caKeyFilePath))

	return nil
}
