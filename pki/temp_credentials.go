package pki

import (
	"fmt"
	"log"
)

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
		publicKey:  privateKey.GetPublicKey(),
	}

	return credentials, nil
}

func (t *TempCredentials) toPermanentCredentials(password []byte, cert *Certificate, filename string, path ...string) (*PermanentCredentials, error) {
	if !cert.GetPublicKey().Equal(t.publicKey) {
		return nil, fmt.Errorf("public key of certificate does not match")
	}

	credentials := getCredentials(password, filename, path...)
	log.Printf("credentials: %+v", credentials)
	err := credentials.Set(cert, t.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to set credentials: %w", err)
	}

	return credentials, nil
}

func (t *TempCredentials) GetPublicKey() *PublicKey {
	return t.publicKey
}

func (t *TempCredentials) GetPrivateKey() *PrivateKey {
	return t.privateKey
}
