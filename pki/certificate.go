package pki

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
)

type Certificate x509.Certificate

func (cert *Certificate) MarshalJSON() ([]byte, error) {
	return json.Marshal(cert.BinaryEncode())
}

func (cert *Certificate) UnmarshalJSON(data []byte) error {
	certBytes := make([]byte, 0, len(data))
	err := json.Unmarshal(data, &certBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal certificate: %w", err)
	}

	newCert, err := CertificateFromBinary(certBytes)
	if err != nil {
		return fmt.Errorf("failed to decode certificate: %w", err)
	}

	*cert = *newCert

	return nil
}

func (cert *Certificate) BinaryEncode() []byte {
	return cert.Raw
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

func (cert *Certificate) ToX509() *x509.Certificate {
	certTyped := x509.Certificate(*cert)
	return &certTyped
}

func ImportCertificate(cert *x509.Certificate) (*Certificate, error) {
	_, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not of type *ecdsa.PublicKey")
	}
	certTyped := Certificate(*cert)
	return &certTyped, nil
}

func (cert *Certificate) GetPublicKey() *PublicKey {
	certTyped, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("public key is not of type *ecdsa.PublicKey")
	}
	pub := PublicKey(*certTyped)
	return &pub
}

func (cert *Certificate) GetType() (CertType, error) {
	if len(cert.Subject.OrganizationalUnit) == 0 {
		return CertTypeError, fmt.Errorf("organizational unit is empty")
	}
	ou := cert.Subject.OrganizationalUnit[0]
	ct := CertType(ou)

	switch ct {
	case CertTypeUser:
		if !cert.IsCA {
			return CertTypeError, fmt.Errorf("user certificate is not a CA")
		}

	}

	return CertTypeError, fmt.Errorf("unknown certificate type")

}

func (cert *Certificate) GetName() string {
	return cert.Subject.CommonName
}
