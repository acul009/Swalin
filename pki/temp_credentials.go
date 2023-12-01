package pki

import (
	"fmt"
)

var _ Credentials = (*TempCredentials)(nil)

type TempCredentials struct {
	publicKey  *PublicKey
	privateKey *PrivateKey
}

func GenerateCredentials() (*TempCredentials, error) {
	privateKey, err := generateKeypair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate keypair: %w", err)
	}

	credentials := &TempCredentials{
		privateKey: privateKey,
		publicKey:  privateKey.PublicKey(),
	}

	return credentials, nil
}

func (t *TempCredentials) ToPermanentCredentials(cert *Certificate) (*PermanentCredentials, error) {
	if !cert.PublicKey().Equal(t.publicKey) {
		return nil, fmt.Errorf("public key of certificate does not match")
	}

	credentials := &PermanentCredentials{
		cert: cert,
		key:  t.privateKey,
	}

	return credentials, nil
}

func (t *TempCredentials) PublicKey() *PublicKey {
	return t.publicKey
}

func (t *TempCredentials) PrivateKey() *PrivateKey {
	return t.privateKey
}
