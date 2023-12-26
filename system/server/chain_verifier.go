package server

import (
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/rahn-it/svalin/pki"
)

var _ pki.Verifier = (*chainVerifier)(nil)

type chainVerifier struct {
	root          *pki.Certificate
	rootPool      *x509.CertPool
	intermediates *x509.CertPool
}

func newChainVerifier(root *pki.Certificate) (*chainVerifier, error) {
	rootPool := x509.NewCertPool()
	rootPool.AddCert(root.ToX509())

	intermediates := x509.NewCertPool()

	return &chainVerifier{
		root:          root,
		rootPool:      rootPool,
		intermediates: intermediates,
	}, nil
}

func (v *chainVerifier) Verify(cert *pki.Certificate) ([]*pki.Certificate, error) {
	if cert.Equal(v.root) {
		return []*pki.Certificate{v.root}, nil
	}

	chain, err := cert.VerifyChain(v.rootPool, v.intermediates)
	if err != nil {
		return nil, fmt.Errorf("failed to verify certificate: %w", err)
	}

	certType := cert.Type()
	if certType == pki.CertTypeError {
		return nil, fmt.Errorf("invalid certificate type: %s", certType)
	}

	return chain, nil
}

func (v *chainVerifier) VerifyPublicKey(pub *pki.PublicKey) ([]*pki.Certificate, error) {
	return nil, errors.New("this verifier is not meant to be used for public keys")
}
