package system

import (
	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/util"
)

type loginParameterRequest struct {
	Username string
}

type loginParameters struct {
	PasswordParams util.ArgonParameters
}

type loginRequest struct {
	PasswordHash []byte
	Totp         string
}

type loginSuccessResponse struct {
	RootCert            *pki.Certificate
	UpstreamCert        *pki.Certificate
	Cert                *pki.Certificate
	EncryptedPrivateKey []byte
}
