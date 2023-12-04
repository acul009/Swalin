package pki

import (
	"crypto/x509"
	"fmt"
)

type Verifier interface {
	Verify(cert *Certificate) ([]*Certificate, error)
	VerifyPublicKey(pub *PublicKey) ([]*Certificate, error)
}

func (c *Certificate) VerifyChain(roots *x509.CertPool, intermediates *x509.CertPool) ([]*Certificate, error) {

	if roots == nil {
		return nil, fmt.Errorf("roots are nil")
	}

	chains, err := c.ToX509().Verify(x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
	})
	if err != nil || len(chains) == 0 {
		return nil, fmt.Errorf("failed to verify certificate: %w", err)
	}

	chain := make([]*Certificate, 0, len(chains[0]))

	for _, cert := range chains[0] {
		workingCert, err := ImportCertificate(cert)
		if err != nil {
			return nil, fmt.Errorf("failed to re-import certificate: %w", err)
		}

		chain = append(chain, workingCert)
	}

	return chain, nil
}
