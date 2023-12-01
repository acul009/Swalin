package pki

import (
	"crypto/tls"
	"encoding/pem"
	"fmt"
)

type Credentials interface {
	GetPrivateKey() (*PrivateKey, error)
	GetPublicKey() (*PublicKey, error)
}

type PermanentCredentials struct {
	cert *Certificate
	key  *PrivateKey
}

func (u *PermanentCredentials) Certificate() *Certificate {
	return u.cert
}

func (u *PermanentCredentials) GetPublicKey() *PublicKey {
	return u.cert.GetPublicKey()
}

func (u *PermanentCredentials) PrivateKey() *PrivateKey {
	return u.key
}

func (u *PermanentCredentials) Get() (*Certificate, *PrivateKey) {
	return u.cert, u.key
}

func (u *PermanentCredentials) GetTlsCert() (*tls.Certificate, error) {
	cert := u.Certificate()

	key := u.PrivateKey()

	tlsCert := &tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  key.ToEcdsa(),
	}

	return tlsCert, nil
}

func (u *PermanentCredentials) GetName() string {
	return u.cert.GetName()
}

func (u *PermanentCredentials) PemEncode(password []byte) ([]byte, error) {
	certPem := u.Certificate().PemEncode()
	keyPem, err := u.PrivateKey().PemEncode(password)
	if err != nil {
		return nil, fmt.Errorf("failed to encode private key: %w", err)
	}

	return append(certPem, keyPem...), nil
}

func CredentialsFromPem(pemBytes []byte, password []byte) (*PermanentCredentials, error) {
	certBlock, rest := pem.Decode(pemBytes)
	if certBlock == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := CertificateFromBinary(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	key, err := PrivateKeyFromPem(rest, password)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &PermanentCredentials{
		cert: cert,
		key:  key,
	}, nil
}
