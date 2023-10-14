package pki

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"time"
)

var CurveToUse = elliptic.P521()

const rootValidFor = 10 * 365 * 24 * time.Hour
const userValidFor = 10 * 365 * 24 * time.Hour
const serverValidFor = 10 * 365 * 24 * time.Hour
const agentValidFor = 2 * 365 * 24 * time.Hour

func generateKeypair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
}

const passwordLength = 64
const passwordCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?"
const charsetLength = len(passwordCharset)

func generatePassword() ([]byte, error) {
	// Create a byte slice to hold the random password
	password := make([]byte, passwordLength)

	max := big.NewInt(int64(charsetLength))

	for i := 0; i < passwordLength; i++ {
		// Generate a random index within the character set length
		index, err := rand.Int(rand.Reader, max)
		if err != nil {
			return []byte{}, err
		}

		// Use the index to select a character from the character set
		password[i] = passwordCharset[index.Int64()]
	}

	return password, nil
}

func generateSerialNumber() (*big.Int, error) {
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	return serialNumber, nil
}

func getTemplate(pub *ecdsa.PublicKey) (*x509.Certificate, error) {
	serial, err := generateSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}
	return &x509.Certificate{
		PublicKey:             pub,
		SerialNumber:          serial,
		SignatureAlgorithm:    x509.ECDSAWithSHA512,
		NotBefore:             time.Now(),
		IsCA:                  false,
		BasicConstraintsValid: true,
	}, nil
}

func generateRootCert(commonName string) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	// Generate a new CA private key
	caPrivateKey, err := generateKeypair()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate CA private key: %w", err)
	}

	caTemplate, err := getTemplate(&caPrivateKey.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate CA template: %w", err)
	}

	caTemplate.Subject = pkix.Name{
		OrganizationalUnit: []string{string(CertTypeRoot)},
		CommonName:         commonName,
	}
	caTemplate.NotAfter = time.Now().Add(rootValidFor)
	caTemplate.KeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	caTemplate.IsCA = true
	caTemplate.BasicConstraintsValid = true

	// Create and save the self-signed CA certificate
	caCert, err := signCert(caTemplate, caPrivateKey, caTemplate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to sign CA certificate: %w", err)
	}

	return caCert, caPrivateKey, nil
}

func signCert(template *x509.Certificate, caKey *ecdsa.PrivateKey, caCert *x509.Certificate) (*x509.Certificate, error) {
	certDER, err := x509.CreateCertificate(rand.Reader, template, caCert, template.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

type CertType string

const (
	CertTypeRoot   CertType = "root"
	CertTypeUser   CertType = "users"
	CertTypeServer CertType = "servers"
	CertTypeAgent  CertType = "agents"
)

func generateUserCert(username string, caKey *ecdsa.PrivateKey, caCert *x509.Certificate) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	// Generate a new user private key
	userPrivateKey, err := generateKeypair()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate user private key: %w", err)
	}

	userTemplate, err := getTemplate(&userPrivateKey.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate CA template: %w", err)
	}

	userTemplate.Subject = pkix.Name{
		OrganizationalUnit: []string{string(CertTypeUser)},
		CommonName:         username,
	}

	userTemplate.NotAfter = time.Now().Add(userValidFor)
	userTemplate.KeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	userTemplate.IsCA = true

	cert, err := signCert(userTemplate, caKey, caCert)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to sign user certificate: %w", err)
	}

	return cert, userPrivateKey, nil
}

func createServerCert(name string, pub *ecdsa.PublicKey, caKey *ecdsa.PrivateKey, caCert *x509.Certificate) (*x509.Certificate, error) {
	serverTemplate, err := getTemplate(pub)
	if err != nil {
		return nil, fmt.Errorf("failed to generate server template: %w", err)
	}

	serverTemplate.Subject = pkix.Name{
		OrganizationalUnit: []string{string(CertTypeServer)},
		CommonName:         name,
	}

	serverTemplate.NotAfter = time.Now().Add(serverValidFor)
	serverTemplate.KeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature

	cert, err := signCert(serverTemplate, caKey, caCert)
	if err != nil {
		return nil, fmt.Errorf("failed to sign server certificate: %w", err)
	}

	return cert, nil
}
