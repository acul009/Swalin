package server

import (
	"bytes"
	"crypto/x509"
	"fmt"

	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/system"
)

var _ pki.Verifier = (*LocalCertificateVerifier)(nil)

type LocalCertificateVerifier struct {
	rootPool        *x509.CertPool
	root            *pki.Certificate
	intermediates   *x509.CertPool
	userStore       *userStore
	deviceStore     *deviceStore
	revocationStore *system.RevocationStore
}

func newLocalCertificateVerifier(root *pki.Certificate, userStore *userStore, deviceStore *deviceStore, revocationStore *system.RevocationStore) (*LocalCertificateVerifier, error) {
	if root == nil {
		return nil, fmt.Errorf("root certificate cannot be nil")
	}

	if userStore == nil {
		return nil, fmt.Errorf("user store cannot be nil")
	}

	if deviceStore == nil {
		return nil, fmt.Errorf("device store cannot be nil")
	}

	if revocationStore == nil {
		return nil, fmt.Errorf("revocation store cannot be nil")
	}

	rootPool := x509.NewCertPool()
	rootPool.AddCert(root.ToX509())

	intermediates := x509.NewCertPool()

	err := userStore.forEach(func(user *user) error {
		intermediates.AddCert(user.Certificate.ToX509())
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add intermediates: %w", err)
	}

	return &LocalCertificateVerifier{
		rootPool:        rootPool,
		root:            root,
		intermediates:   intermediates,
		userStore:       userStore,
		deviceStore:     deviceStore,
		revocationStore: revocationStore,
	}, nil
}

func (v *LocalCertificateVerifier) Verify(cert *pki.Certificate) ([]*pki.Certificate, error) {
	if cert.Equal(v.root) {
		return []*pki.Certificate{v.root}, nil
	}

	chain, err := v.verifyChain(cert)
	if err != nil {
		return nil, err
	}

	knownCert, err := v.findCertificate(cert.PublicKey())
	if err != nil {
		return nil, fmt.Errorf("failed to find given certificate: %w", err)
	}

	if !bytes.Equal(knownCert.BinaryEncode(), cert.BinaryEncode()) {
		return nil, fmt.Errorf("certificate does not match known certificate for the given public key")
	}

	return chain, nil
}

func (v *LocalCertificateVerifier) VerifyPublicKey(pub *pki.PublicKey) ([]*pki.Certificate, error) {
	if pub.Equal(v.root.PublicKey()) {
		return []*pki.Certificate{v.root}, nil
	}

	cert, err := v.findCertificate(pub)
	if err != nil {
		return nil, err
	}

	return v.verifyChain(cert)
}

func (v *LocalCertificateVerifier) findCertificate(pub *pki.PublicKey) (*pki.Certificate, error) {
	user, err := v.userStore.getUser(pub)
	if err != nil {
		return nil, fmt.Errorf("failed to check for user: %w", err)
	}

	if user != nil {
		return user.Certificate, nil
	}

	device, err := v.deviceStore.GetDevice(pub)
	if err != nil {
		return nil, fmt.Errorf("failed to check for device: %w", err)
	}

	if device != nil {
		return device, nil
	}

	return nil, fmt.Errorf("the public key does not match any known user or device")
}

func (v *LocalCertificateVerifier) verifyChain(cert *pki.Certificate) ([]*pki.Certificate, error) {
	chain, err := cert.VerifyChain(v.rootPool, v.intermediates)
	if err != nil {
		return nil, fmt.Errorf("failed to verify certificate: %w", err)
	}

	for _, c := range chain {
		err := v.revocationStore.CheckCertificate(c)
		if err != nil {
			return nil, fmt.Errorf("certificate in chain is revoked: %w", err)
		}
	}

	return chain, nil
}
