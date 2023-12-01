package pki

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding"
	"encoding/pem"
	"fmt"
	"log"
)

var _ encoding.TextUnmarshaler = (*Certificate)(nil)

type Certificate struct {
	cert *x509.Certificate
}

func (cert *Certificate) MarshalText() ([]byte, error) {
	return cert.PemEncode(), nil
}

func (cert *Certificate) UnmarshalText(data []byte) error {
	newCert, err := CertificateFromPem(data)
	if err != nil {
		return fmt.Errorf("failed to decode certificate: %w", err)
	}

	*cert = *newCert
	return nil
}

func (c *Certificate) BinaryEncode() []byte {
	return c.cert.Raw
}

func CertificateFromBinary(cert []byte) (*Certificate, error) {
	certTyped, err := x509.ParseCertificate(cert)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}
	return ImportCertificate(certTyped)
}

func (cert *Certificate) PemEncode() []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
}

func CertificateFromPem(certPEM []byte) (*Certificate, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	return CertificateFromBinary(block.Bytes)
}

func (c *Certificate) ToX509() *x509.Certificate {
	return c.cert
}

func ImportCertificate(cert *x509.Certificate) (*Certificate, error) {
	_, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not of type *ecdsa.PublicKey")
	}

	pkiCert := &Certificate{
		cert: cert,
	}

	return pkiCert, nil
}

func (c *Certificate) Equal(compare *Certificate) bool {
	return bytes.Equal(c.cert.Raw, compare.cert.Raw)
}

func (c *Certificate) PublicKey() *PublicKey {
	keyTyped, ok := c.cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("public key is not of type *ecdsa.PublicKey")
	}
	pub := PublicKey(*keyTyped)
	return &pub
}

func (c *Certificate) GetName() string {
	return c.cert.Subject.CommonName
}

func (c *Certificate) Type() CertType {
	if len(c.cert.Subject.OrganizationalUnit) == 0 {
		return CertTypeError
	}

	t := CertType(c.cert.Subject.OrganizationalUnit[0])

	if t == CertTypeUser || t == CertTypeRoot {
		if !c.cert.IsCA {
			log.Printf("WARNING: certificate of type %s is not a CA", t)
			return CertTypeError
		}

		return t
	}

	if t == CertTypeAgent || t == CertTypeServer {

		if c.cert.IsCA {
			log.Printf("WARNING: certificate of type %s is a CA", t)
			return CertTypeError
		}

		return t
	}

	return CertTypeError
}
