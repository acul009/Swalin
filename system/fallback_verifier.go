package system

import (
	"crypto/x509"
	"fmt"

	"github.com/rahn-it/svalin/pki"
)

var _ pki.Verifier = (*fallbackVerifier)(nil)

type fallbackVerifier struct {
	root     *pki.Certificate
	upstream *pki.Certificate
}

func NewFallbackVerifier(root *pki.Certificate, upstream *pki.Certificate) (*fallbackVerifier, error) {
	rootPool := x509.NewCertPool()
	rootPool.AddCert(root.ToX509())

	intermediates := x509.NewCertPool()

	_, err := upstream.VerifyChain(rootPool, intermediates)
	if err != nil {
		return nil, fmt.Errorf("failed to verify upstream certificate: %w", err)
	}

	return &fallbackVerifier{
		root:     root,
		upstream: upstream,
	}, nil
}

func (v *fallbackVerifier) Verify(cert *pki.Certificate) ([]*pki.Certificate, error) {
	if cert.Equal(v.root) {
		return []*pki.Certificate{v.root}, nil
	}

	if cert.Equal(v.upstream) {
		return []*pki.Certificate{v.upstream, v.root}, nil
	}

	return nil, fmt.Errorf("failed to verify certificate via fallback")
}

func (v *fallbackVerifier) VerifyPublicKey(pub *pki.PublicKey) ([]*pki.Certificate, error) {
	if v.root.PublicKey().Equal(pub) {
		return []*pki.Certificate{v.root}, nil
	}

	if v.upstream.PublicKey().Equal(pub) {
		return []*pki.Certificate{v.upstream, v.root}, nil
	}

	return nil, fmt.Errorf("failed to verify public key via fallback")
}
