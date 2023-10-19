package pki

import (
	"bytes"
	"crypto/x509"
	"fmt"
)

func VerifyCertificate(cert *x509.Certificate) error {
	if cert == nil {
		return fmt.Errorf("certificate is nil")
	}

	rootCert, err := GetRootCert()
	if err != nil {
		return fmt.Errorf("failed to load root certificate: %w", err)
	}

	// certificate is not root

	if bytes.Equal(rootCert.Raw, cert.Raw) {
		return nil
	}

	upstreamCert, err := GetUpstreamCert()
	if err != nil {
		return fmt.Errorf("failed to load upstream certificate: %w", err)
	}

	//TODO: actual certificate verification

	if bytes.Equal(upstreamCert.Raw, cert.Raw) {
		return nil
	}

	return fmt.Errorf("certificate is not known")
}

func VerifyUserCertificate(cert *x509.Certificate, username string) error {
	if cert == nil {
		return fmt.Errorf("certificate is nil")
	}

	err := VerifyCertificate(cert)
	if err != nil {
		return fmt.Errorf("failed to verify certificate: %w", err)
	}

	if cert.Subject.OrganizationalUnit[0] != string(CertTypeUser) {
		return fmt.Errorf("certificate is not a user certificate")
	}

	if cert.Subject.CommonName != username {
		return fmt.Errorf("certificate is not for user %s", username)
	}

	if !cert.IsCA {
		return fmt.Errorf("certificate is not a CA")
	}

	return nil
}
