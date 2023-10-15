package pki

import (
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"rahnit-rmm/config"
)

type NotUnlockedError struct {
}

func (e NotUnlockedError) Error() string {
	return "current key is not unlocked"
}

func (e NotUnlockedError) Is(target error) bool {
	_, ok := target.(NotUnlockedError)
	return ok
}

var currentKey *ecdsa.PrivateKey = nil
var currentPub *ecdsa.PublicKey = nil
var currentCert *x509.Certificate = nil

const currentKeyFilePath = "current.key"
const currentCertFilePath = "current.crt"
const currentPubFilePath = "current.pub"

func CurrentAvailable() (bool, error) {
	_, err := os.Stat(config.GetFilePath(currentCertFilePath))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check if current cert exists: %w", err)
	}
	return true, nil
}

func CurrentPublicKeyAvailable() (bool, error) {
	_, err := os.Stat(config.GetFilePath(currentPubFilePath))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check if current public key exists: %w", err)
	}
	return true, nil
}

func CurrentAvailableUser() (string, error) {
	cert, err := LoadCertFromFile(config.GetFilePath(currentCertFilePath))
	if err != nil {
		return "", fmt.Errorf("failed to load current cert: %w", err)
	}

	// check if the OU is users
	ou := cert.Subject.OrganizationalUnit[0]
	if ou != string(CertTypeUser) {
		return "", fmt.Errorf("current cert is not a user certificate")
	}

	return cert.Subject.CommonName, nil
}

func Unlock(password []byte) error {
	cert, err := LoadCertFromFile(config.GetFilePath(currentCertFilePath))
	var pub *ecdsa.PublicKey
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("failed to load current cert: %w", err)
		}

		pub, err = LoadPublicKeyFromFile(config.GetFilePath(currentPubFilePath))
		if err != nil {
			return fmt.Errorf("failed to load current public key: %w", err)
		}
	} else {
		var ok bool
		pub, ok = cert.PublicKey.(*ecdsa.PublicKey)
		if !ok {
			return fmt.Errorf("public key is not of type *ecdsa.PublicKey")
		}
	}

	key, err := LoadCertKeyFromFile(config.GetFilePath(currentKeyFilePath), password)
	if err != nil {
		return fmt.Errorf("failed to load current key: %w", err)
	}

	currentCert = cert
	currentPub = pub
	currentKey = key

	return nil
}

func UnlockAsRoot(password []byte) error {
	caCert, caKey, err := GetRoot(password)
	if err != nil {
		return fmt.Errorf("failed to load CA: %w", err)
	}

	pub, ok := currentCert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("public key is not of type *ecdsa.PublicKey")
	}

	currentKey = caKey
	currentCert = caCert
	currentPub = pub

	return nil
}

func GetCurrentKey() (*ecdsa.PrivateKey, error) {
	if currentKey == nil {
		return nil, NotUnlockedError{}
	}
	return currentKey, nil
}

func GetCurrentCert() (*x509.Certificate, error) {
	if currentCert == nil {
		return nil, NotUnlockedError{}
	}
	return currentCert, nil
}

func GetCurrentPublicKey() (*ecdsa.PublicKey, error) {
	if currentPub == nil {
		return nil, NotUnlockedError{}
	}
	return currentPub, nil
}

func GetCurrentTlsCert() (*tls.Certificate, error) {
	cert, err := GetCurrentCert()
	if err != nil {
		return nil, fmt.Errorf("failed to get current cert: %w", err)
	}

	tlsCert := &tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  currentKey,
	}

	return tlsCert, nil
}

func SaveCurrentCertAndKey(cert *x509.Certificate, key *ecdsa.PrivateKey, password []byte) error {
	err := SaveCurrentCert(cert)
	if err != nil {
		return fmt.Errorf("failed to save current cert: %w", err)
	}
	err = SaveCertKeyToFile(config.GetFilePath(currentKeyFilePath), key, password)
	if err != nil {
		return fmt.Errorf("failed to save current key: %w", err)
	}
	return nil
}

func SaveCurrentKeyPair(key *ecdsa.PrivateKey, pub *ecdsa.PublicKey, password []byte) error {
	err := SaveCertKeyToFile(config.GetFilePath(currentKeyFilePath), key, password)
	if err != nil {
		return fmt.Errorf("failed to save current key: %w", err)
	}
	err = SavePublicKeyToFile(config.GetFilePath(currentPubFilePath), pub)
	if err != nil {
		return fmt.Errorf("failed to save current public key: %w", err)
	}
	return nil
}

func SaveCurrentCert(cert *x509.Certificate) error {
	err := SaveCertToFile(config.GetFilePath(currentCertFilePath), cert)
	if err != nil {
		return fmt.Errorf("failed to save current cert: %w", err)
	}
	return nil
}
