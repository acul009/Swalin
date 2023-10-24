package pki

import (
	"bytes"
	"fmt"
)

func VerifyCertificate(cert *Certificate) error {
	if cert == nil {
		return fmt.Errorf("certificate is nil")
	}

	rootCert, err := GetRootCert()
	if err != nil {
		return fmt.Errorf("failed to load root certificate: %w", err)
	}

	if bytes.Equal(rootCert.Raw, cert.Raw) {
		return nil
	}

	// certificate is not root

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

func VerifyUserCertificate(cert *Certificate) error {
	if cert == nil {
		return fmt.Errorf("certificate is nil")
	}

	err := VerifyCertificate(cert)
	if err != nil {
		return fmt.Errorf("failed to verify certificate: %w", err)
	}

	if cert.Subject.OrganizationalUnit[0] != string(CertTypeUser) && cert.Subject.OrganizationalUnit[0] != string(CertTypeRoot) {
		return fmt.Errorf("certificate is not a user certificate")
	}

	if !cert.IsCA {
		return fmt.Errorf("certificate is not a CA")
	}

	return nil
}
