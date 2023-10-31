package pki

import (
	"context"
	"crypto/x509"
	"fmt"
	"rahnit-rmm/config"
)

type verifyOptionCache struct {
	rootPool      *x509.CertPool
	intermediates *x509.CertPool
}

func NewOptionCacheFromDB() (*verifyOptionCache, error) {
	rootCert, err := Root.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to load root certificate: %w", err)
	}

	rootPool := x509.NewCertPool()
	rootPool.AddCert(rootCert.ToX509())

	intermediatePool := x509.NewCertPool()

	db := config.DB()
	users, err := db.User.Query().All(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	for _, user := range users {
		userCert, err := CertificateFromPem([]byte(user.Certificate))
		if err != nil {
			return nil, fmt.Errorf("failed to load user certificate: %w", err)
		}

		intermediatePool.AddCert(userCert.ToX509())
	}

	return &verifyOptionCache{
		rootPool:      rootPool,
		intermediates: intermediatePool,
	}, nil
}

func (v *verifyOptionCache) options() x509.VerifyOptions {

	return x509.VerifyOptions{
		Roots:         v.rootPool,
		Intermediates: v.intermediates,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
}

func (v *verifyOptionCache) verify(cert *Certificate) error {
	chains, err := cert.ToX509().Verify(v.options())
	if err != nil {
		return fmt.Errorf("failed to verify certificate: %w", err)
	}

	for _, cert := range chains[0] {
		workingCert, err := ImportCertificate(cert)
		if err != nil {
			return fmt.Errorf("failed to import certificate: %w", err)
		}

		err = v.checkCertificateInfo(workingCert)
		if err != nil {
			return fmt.Errorf("failed to check certificate info: %w", err)
		}
	}

	return nil
}

func (v *verifyOptionCache) checkCertificateInfo(cert *Certificate) error {
	err := v.checkRevoked(cert)
	if err != nil {
		return fmt.Errorf("certificate has been revoked: %w", err)
	}

	return nil
}

func (v *verifyOptionCache) checkRevoked(cert *Certificate) error {
	// TODO
	return nil
}

func (v *verifyOptionCache) VerifyUser(cert *Certificate) error {
	if cert == nil {
		return fmt.Errorf("certificate is nil")
	}

	err := v.verify(cert)
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

func (v *verifyOptionCache) VerifyAgent(cert *Certificate) error {
	if cert == nil {
		return fmt.Errorf("certificate is nil")
	}

	err := v.verify(cert)
	if err != nil {
		return fmt.Errorf("failed to verify certificate: %w", err)
	}

	if cert.Subject.OrganizationalUnit[0] != string(CertTypeAgent) {
		return fmt.Errorf("certificate is not an agent certificate")
	}

	if cert.IsCA {
		return fmt.Errorf("certificate is a CA")
	}

	return nil
}
