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

const rootValidFor = 10 * 365 * 24 * time.Hour
const userValidFor = 10 * 365 * 24 * time.Hour

func generateKeypair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
}

func generateSerialNumber() (*big.Int, error) {
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %v", err)
	}

	return serialNumber, nil
}

func getTemplate(pub *ecdsa.PublicKey) (*x509.Certificate, error) {
	serial, err := generateSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %v", err)
	}
	return &x509.Certificate{
		PublicKey:          pub,
		SerialNumber:       serial,
		SignatureAlgorithm: x509.ECDSAWithSHA512,
		NotBefore:          time.Now(),
	}, nil
}

func generateRootCert(commonName string) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	// Generate a new CA private key
	caPrivateKey, err := generateKeypair()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate CA private key: %v", err)
	}

	caTemplate, err := getTemplate(&caPrivateKey.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate CA template: %v", err)
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
		return nil, nil, fmt.Errorf("failed to sign CA certificate: %v", err)
	}

	return caCert, caPrivateKey, nil
}

func signCert(template *x509.Certificate, caKey *ecdsa.PrivateKey, caCert *x509.Certificate) (*x509.Certificate, error) {
	certDER, err := x509.CreateCertificate(rand.Reader, template, caCert, template.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return cert, nil
}

type CertType string

const (
	CertTypeRoot CertType = "root"
	CertTypeUser CertType = "users"
)

func generateUserCert(username string, caKey *ecdsa.PrivateKey, caCert *x509.Certificate) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	// Generate a new user private key
	userPrivateKey, err := generateKeypair()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate user private key: %v", err)
	}

	userTemplate, err := getTemplate(&userPrivateKey.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate CA template: %v", err)
	}

	userTemplate.Subject = pkix.Name{
		OrganizationalUnit: []string{string(CertTypeUser)},
		CommonName:         username,
	}

	userTemplate.NotAfter = time.Now().Add(userValidFor)
	userTemplate.KeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	userTemplate.IsCA = true
	userTemplate.BasicConstraintsValid = true

	cert, err := signCert(userTemplate, caKey, caCert)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to sign user certificate: %v", err)
	}

	return cert, userPrivateKey, nil
}
