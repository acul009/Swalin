package rpc

import (
	"fmt"
	"rahnit-rmm/config"
	"rahnit-rmm/pki"
	"rahnit-rmm/util"

	"github.com/pquerna/otp/totp"
)

func RegisterUserHandler() RpcCommand {
	return &registerUserCmd{}
}

type registerUserCmd struct {
	Username             string
	Cert                 []byte
	EncryptedPrivateKey  []byte
	ClientHashingOptions util.ArgonParameters
	PasswordHash         []byte
	TotpSecret           string
	CurrentTotp          string
}

func (r *registerUserCmd) ExecuteServer(session *RpcSession) error {
	// check if hashing options are ok
	if r.ClientHashingOptions.IsInsecure() {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 400,
			Msg:  "Insecure Argon Parameters",
		})
		return fmt.Errorf("insecure Argon Parameters")
	}

	// check if certificate is ok
	cert, err := pki.DecodeCertificate(r.Cert)
	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 400,
			Msg:  "Invalid certificate",
		})
		return fmt.Errorf("invalid certificate: %w", err)
	}

	// check if totp secret is ok and current totp is valid
	if !totp.Validate(r.TotpSecret, r.CurrentTotp) {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 400,
			Msg:  "Invalid TOTP",
		})
		return fmt.Errorf("invalid TOTP")
	}

	err = pki.VerifyUserCertificate(cert, r.Username)
	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 400,
			Msg:  "Invalid certificate",
		})
		return fmt.Errorf("invalid certificate: %w", err)
	}

	// Request seems valid, hash the password again
	hashingOpts, err := util.GenerateArgonParameters()
	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 500,
			Msg:  "failed to generate Argon Parameters",
		})
		return fmt.Errorf("failed to generate Argon Parameters: %w", err)
	}

	double_hash, err := util.HashPassword(r.PasswordHash, hashingOpts)
	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 500,
			Msg:  "failed to hash password",
		})
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// create user
	db := config.DB()

	err = db.User.Create().
		SetUsername(r.Username).
		SetCertificate(string(r.Cert)).
		SetEncryptedPrivateKey(string(r.EncryptedPrivateKey)).
		SetPasswordClientHashingOptions(&r.ClientHashingOptions).
		SetPasswordServerHashingOptions(&hashingOpts).
		SetPasswordDoubleHashed(string(double_hash)).
		SetTotpSecret(r.TotpSecret).
		Exec(session.Context())

	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 500,
			Msg:  "failed to create user",
		})
		return fmt.Errorf("failed to create user: %w", err)
	}

	session.WriteResponseHeader(SessionResponseHeader{
		Code: 200,
		Msg:  "User created",
	})
	return nil
}

func (r *registerUserCmd) ExecuteClient(session *RpcSession) error {
	return nil
}

func (r *registerUserCmd) GetKey() string {
	return "register-user"
}
