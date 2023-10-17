package pki

import (
	"crypto/x509"
	"fmt"
	"rahnit-rmm/config"
)

const upstremCertFile = "upstream.crt"

var upstream *x509.Certificate

func GetUpstreamCert() (*x509.Certificate, error) {
	if upstream == nil {
		var err error
		upstream, err = LoadCertFromFile(config.GetFilePath(upstremCertFile))
		if err != nil {
			return nil, fmt.Errorf("failed to load upstream certificate: %w", err)
		}
	}
	return upstream, nil
}

func SaveUpstreamCert(cert *x509.Certificate) error {
	if cert == nil {
		return fmt.Errorf("certificate is nil")
	}

	err := SaveCertToFile(config.GetFilePath(upstremCertFile), cert)
	if err != nil {
		return fmt.Errorf("failed to save upstream certificate: %w", err)
	}

	upstream = cert
	return nil
}
