package pki

import (
	"crypto/x509"
	"fmt"
)

type Verifier interface {
	Verify(cert *Certificate) ([]*Certificate, error)
	VerifyPublicKey(pub *PublicKey) ([]*Certificate, error)
}

func (c *Certificate) VerifyChain(roots *x509.CertPool, intermediates *x509.CertPool, checkRevocation bool) ([]*Certificate, error) {

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

		if checkRevocation {
			err = workingCert.checkRevoked()
			if err != nil {
				return nil, fmt.Errorf("failed to check revocation: %w", err)
			}
		}

	}

	return chain, nil
}

func (c *Certificate) checkRevoked() error {
	// TODO
	return nil
}

func CreatePool(certs []*Certificate) *x509.CertPool {
	pool := x509.NewCertPool()

	for _, cert := range certs {
		pool.AddCert(cert.ToX509())
	}

	return pool
}
