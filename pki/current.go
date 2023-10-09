package pki

import (
	"crypto/ecdsa"
	"crypto/x509"
	"fmt"
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
var currentCert *x509.Certificate = nil

const currentKeyFilePath = "current.key"
const currentCertFilePath = "current.cert"

func Unlock(password []byte) error {
	cert, err := LoadCertFromFile(config.GetFilePath(currentCertFilePath))
	if err != nil {
		return fmt.Errorf("failed to load current cert: %v", err)
	}

	key, err := LoadCertKeyFromFile(config.GetFilePath(currentKeyFilePath), password)
	if err != nil {
		return fmt.Errorf("failed to load current key: %v", err)
	}
	currentCert = cert
	currentKey = key
	return nil
}

func UnlockAsRoot(password []byte) error {
	caCert, caKey, err := GetCa(password)
	if err != nil {
		return fmt.Errorf("failed to load CA: %v", err)
	}

	currentKey = caKey
	currentCert = caCert

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
	pub, err := GetCurrentCert()
	if err != nil {
		return nil, err
	}
	typed, ok := pub.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not of type *ecdsa.PublicKey")
	}
	return typed, nil
}

func SaveCurrentCertAndKey(cert *x509.Certificate, key *ecdsa.PrivateKey, password []byte) error {
	err := SaveCertToFile(config.GetFilePath(currentCertFilePath), cert)
	if err != nil {
		return fmt.Errorf("failed to save current cert: %v", err)
	}
	err = SaveCertKeyToFile(config.GetFilePath(currentKeyFilePath), key, password)
	if err != nil {
		return fmt.Errorf("failed to save current key: %v", err)
	}
	return nil
}
