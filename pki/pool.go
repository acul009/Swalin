package pki

import (
	"crypto/x509"
	"fmt"
)

type CertPool struct {
	x509.CertPool
}

var rootPool *x509.CertPool = nil

func updateRootPool() {
	rootPool = x509.NewCertPool()
	caCert, err := GetCaCert()
	if err != nil {
		return
	}

	rootPool.AddCert(caCert)
}

func GetRootPool() *x509.CertPool {
	return rootPool
}

func CreateCertPool() *CertPool {
	pool := x509.NewCertPool()

	return &CertPool{CertPool: *pool}
}

func (c *CertPool) Verify(cert *x509.Certificate) error {
	opts := x509.VerifyOptions{
		Roots:         rootPool,
		Intermediates: &c.CertPool,
	}

	//TODO: check for revocations

	_, err := cert.Verify(opts)
	if err != nil {
		return fmt.Errorf("failed to verify certificate: %v", err)
	}

	return nil
}

func (c *CertPool) AddCert(cert *x509.Certificate) error {
	err := c.Verify(cert)
	if err != nil {
		fmt.Println("failed to verify certificate")
	}

	c.CertPool.AddCert(cert)
	return nil
}

func init() {
	updateRootPool()
}
