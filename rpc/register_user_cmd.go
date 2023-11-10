package rpc

import (
	"fmt"
	"rahnit-rmm/config"
	"rahnit-rmm/pki"
	"rahnit-rmm/util"
)

func RegisterUserHandler() RpcCommand {
	return &registerUserCmd{}
}

func NewRegisterUserCmd(credentials *pki.PermanentCredentials, password []byte, totpSecret string, currentTotp string) (*registerUserCmd, error) {
	privateKey, err := credentials.GetPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	encryptedPrivateKey, err := privateKey.BinaryEncode(password)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize private key: %w", err)
	}

	clientHashingParameters, err := util.GenerateArgonParameters(util.ArgonStrengthStrong)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hashing parameters: %w", err)
	}

	passwordHash, err := util.HashPassword(password, clientHashingParameters)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	cert, err := credentials.GetCertificate()
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}

	return &registerUserCmd{
		Cert:                    cert,
		EncryptedPrivateKey:     encryptedPrivateKey,
		ClientHashingParameters: clientHashingParameters,
		PasswordHash:            passwordHash,
		TotpSecret:              totpSecret,
		CurrentTotp:             currentTotp,
	}, nil
}

type registerUserCmd struct {
	Cert                    *pki.Certificate
	EncryptedPrivateKey     []byte
	ClientHashingParameters util.ArgonParameters
	PasswordHash            []byte
	TotpSecret              string
	CurrentTotp             string
}

func (r *registerUserCmd) ExecuteServer(session *RpcSession) error {
	// check if hashing options are ok
	if r.ClientHashingParameters.IsInsecure() {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 400,
			Msg:  "Insecure Argon Parameters",
		})
		return fmt.Errorf("insecure Argon Parameters")
	}

	// check if totp secret is ok and current totp is valid
	if !util.ValidateTotp(r.TotpSecret, r.CurrentTotp) {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 400,
			Msg:  "Invalid TOTP",
		})
		return fmt.Errorf("invalid TOTP")
	}

	cert := r.Cert

	_, err := session.Verifier().Verify(cert)
	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 400,
			Msg:  "Invalid certificate",
		})
		return fmt.Errorf("invalid certificate: %w", err)
	}

	if cert.Type() != pki.CertTypeUser && cert.Type() != pki.CertTypeRoot {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 400,
			Msg:  "Invalid certificate type",
		})
		return fmt.Errorf("invalid certificate type")

	}

	username := cert.Subject.CommonName

	// Request seems valid, hash the password again
	hashingOpts, err := util.GenerateArgonParameters(util.ArgonStrengthDefault)
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

	encodedPub := cert.GetPublicKey().Base64Encode()

	// create user
	db := config.DB()

	err = db.User.Create().
		SetUsername(username).
		SetCertificate(string(cert.PemEncode())).
		SetPublicKey(encodedPub).
		SetEncryptedPrivateKey(r.EncryptedPrivateKey).
		SetPasswordClientHashingOptions(&r.ClientHashingParameters).
		SetPasswordServerHashingOptions(&hashingOpts).
		SetPasswordDoubleHashed(double_hash).
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
