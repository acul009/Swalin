package pki

import (
	"crypto/tls"
	"fmt"
	"path/filepath"
	"rahnit-rmm/config"
)

type Credentials interface {
	GetPrivateKey() (*PrivateKey, error)
	GetPublicKey() (*PublicKey, error)
}

type PermanentCredentials struct {
	certStorage       *storedCertificate
	privateKeyStorage *storedPrivateKey
}

func GetUserCredentials(username string, password []byte) (*PermanentCredentials, error) {
	credentials := getCredentials(password, username, "users")
	if !credentials.Available() {
		return nil, fmt.Errorf("user credentials not found")
	}

	return credentials, nil
}

func SaveUserCredentials(username string, password []byte, cert *Certificate, key *PrivateKey) error {
	credentials := getCredentials(password, username, "users")
	err := credentials.Set(cert, key)
	if err != nil {
		return fmt.Errorf("failed to save user credentials: %w", err)
	}

	return nil
}

func ListAvailableUserCredentials() ([]string, error) {
	userFolder := config.GetFilePath("users", "*.key")

	matches, err := filepath.Glob(userFolder)
	if err != nil {
		return nil, fmt.Errorf("failed to list user credentials: %w", err)
	}

	users := make([]string, 0, len(matches))

	for _, match := range matches {
		newUser := filepath.Base(match)
		ext := filepath.Ext(match)
		users = append(users, newUser[0:len(newUser)-len(ext)])
	}

	return users, nil
}

func getCredentials(password []byte, filename string, path ...string) *PermanentCredentials {
	certStorage := &storedCertificate{
		allowOverride: false,
		filename:      filepath.Join(append(path, filename+".crt")...),
	}

	keyStorage := &storedPrivateKey{
		allowOverride: false,
		filename:      filepath.Join(append(path, filename+".key")...),
		password:      password,
	}

	return &PermanentCredentials{
		certStorage:       certStorage,
		privateKeyStorage: keyStorage,
	}
}

func (u *PermanentCredentials) Available() bool {
	return u.certStorage.Available() && u.privateKeyStorage.Available()
}

func (u *PermanentCredentials) Get() (*Certificate, *PrivateKey, error) {
	cert, err := u.certStorage.Get()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get certificate: %w", err)
	}

	key, err := u.privateKeyStorage.Get()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get private key: %w", err)
	}

	return cert, key, nil
}

func (u *PermanentCredentials) Set(cert *Certificate, key *PrivateKey) error {
	err := u.certStorage.Set(cert)
	if err != nil {
		return fmt.Errorf("failed to set certificate: %w", err)
	}

	err = u.privateKeyStorage.Set(key)
	if err != nil {
		return fmt.Errorf("failed to set private key: %w", err)
	}

	return nil
}

func (u *PermanentCredentials) GetCertificate() (*Certificate, error) {
	return u.certStorage.Get()
}

func (u *PermanentCredentials) GetPublicKey() (*PublicKey, error) {
	cert, err := u.GetCertificate()
	if err != nil {
		return nil, fmt.Errorf("failed to get current cert: %w", err)
	}

	return cert.GetPublicKey(), nil
}

func (u *PermanentCredentials) GetPrivateKey() (*PrivateKey, error) {
	return u.privateKeyStorage.Get()
}

func (u *PermanentCredentials) GetTlsCert() (*tls.Certificate, error) {
	cert, err := u.GetCertificate()
	if err != nil {
		return nil, fmt.Errorf("failed to get current cert: %w", err)
	}

	key, err := u.GetPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get current key: %w", err)
	}

	tlsCert := &tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  key.ToEcdsa(),
	}

	return tlsCert, nil
}
