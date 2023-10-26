package pki

import (
	"fmt"
	"rahnit-rmm/config"
)

const upstremCertFile = "upstream.crt"

var upstream *Certificate

func GetUpstreamCert() (*Certificate, error) {
	if upstream == nil {
		var err error
		upstream, err = loadCertificateFromFile(config.GetFilePath(upstremCertFile))
		if err != nil {
			return nil, fmt.Errorf("failed to load upstream certificate: %w", err)
		}
	}
	return upstream, nil
}

func SaveUpstreamCert(cert *Certificate) error {
	if cert == nil {
		return fmt.Errorf("certificate is nil")
	}

	err := cert.saveToFile(config.GetFilePath(upstremCertFile))
	if err != nil {
		return fmt.Errorf("failed to save upstream certificate: %w", err)
	}

	upstream = cert
	return nil
}
