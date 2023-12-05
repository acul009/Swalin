package system

import (
	"fmt"

	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/rpc"
	"github.com/rahn-it/svalin/util"
)

var _ rpc.RpcCommand = (*registerUserCommand)(nil)

type registerUserCommand struct {
	Certificate         *pki.Certificate
	EncryptedKey        []byte
	PasswordHash        []byte
	TotpSecret          string
	CurrentTotp         string
	ClientHashingParams *util.ArgonParameters
	verifier            pki.Verifier
	acceptUser          func(
		Certificate *pki.Certificate,
		EncryptedPrivateKey []byte,
		ClientHashingParams *util.ArgonParameters,
		ServerHashingParams *util.ArgonParameters,
		DoubleHashedPassword []byte,
		TotpSecret []byte,
	) error
}

func (cmd *registerUserCommand) GetKey() string {
	return "register-user"
}

func (cmd *registerUserCommand) ExecuteClient(session *rpc.RpcSession) error {
	return nil
}

func (cmd *registerUserCommand) ExecuteServer(session *rpc.RpcSession) error {

	if cmd.ClientHashingParams.IsInsecure() {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 400,
			Msg:  "Client hashing parameters are insecure",
		})
		return fmt.Errorf("client hashing parameters are insecure")
	}

	if !util.ValidateTotp(cmd.TotpSecret, cmd.CurrentTotp) {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 400,
			Msg:  "Invalid TOTP code",
		})
		return fmt.Errorf("invalid TOTP code")
	}

	cert := cmd.Certificate

	_, err := cmd.verifier.Verify(cert)
	if err != nil {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 400,
			Msg:  "Invalid certificate",
		})
		return fmt.Errorf("invalid certificate: %w", err)
	}

	if cert.Type() != pki.CertTypeUser && cert.Type() != pki.CertTypeRoot {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 400,
			Msg:  "Invalid certificate type",
		})
		return fmt.Errorf("invalid certificate type")
	}

	serverHashingParams, err := util.GenerateArgonParameters(util.ArgonStrengthDefault)
	if err != nil {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 500,
			Msg:  "failed to generate Argon Parameters",
		})
		return fmt.Errorf("failed to generate Argon Parameters: %w", err)
	}

	double_hash, err := util.HashPassword(cmd.PasswordHash, serverHashingParams)
	if err != nil {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 500,
			Msg:  "failed to hash password",
		})
		return fmt.Errorf("failed to hash password: %w", err)
	}

	err = cmd.acceptUser(cert, cmd.EncryptedKey, cmd.ClientHashingParams, &serverHashingParams, double_hash, []byte(cmd.TotpSecret))
	if err != nil {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 500,
			Msg:  "failed to accept user",
		})
		return fmt.Errorf("failed to accept user: %w", err)
	}

	session.WriteResponseHeader(rpc.SessionResponseHeader{
		Code: 200,
		Msg:  "User accepted",
	})

	return nil
}
