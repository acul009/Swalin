package pki

import (
	"rahnit-rmm/config"
)

type userCredentials struct {
	Username    string
	certStorage *storedCertificate
}

func GetUserCredentials(username string, password []byte) (*userCredentials, error) {
	certStorage := &storedCertificate{
		allowOverride: false,
		filename:      config.GetFilePath("users", username+".crt"),
	}

	return &userCredentials{
		Username:    username,
		certStorage: certStorage,
	}, nil
}

func (u *userCredentials) GetCertificate() (*Certificate, error) {
	return u.certStorage.Get()
}
