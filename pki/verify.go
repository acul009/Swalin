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

	if bytes.Equal(rootCert.Raw, cert.Raw) {
		return nil
	}

	upstreamCert, err := GetUpstreamCert()
	if err != nil {
		return fmt.Errorf("failed to load upstream certificate: %w", err)
	}

	if bytes.Equal(upstreamCert.Raw, cert.Raw) {
		return nil
	}

	return fmt.Errorf("certificate is not known")
}
