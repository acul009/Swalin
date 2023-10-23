package rpc

import (
	"fmt"
	"rahnit-rmm/config"
	"rahnit-rmm/ent"
	"rahnit-rmm/ent/user"
	"rahnit-rmm/pki"
	"rahnit-rmm/util"
)

func Login(addr string, username string, password []byte, totpCode string) error {

}

type loginParameterRequest struct {
	username string
}

type loginParameters struct {
	passwordParams util.ArgonParameters
}

type loginRequest struct {
	passwordHash []byte
	totp         string
}

type loginSuccessResponse struct {
	Cert                string
	EncryptedPrivateKey string
}

func acceptLoginRequest(conn *rpcConnection) error {
	// prepare session
	ctx := conn.connection.Context()
	defer conn.Close(500, "")

	session, err := conn.OpenSession(ctx)
	if err != nil {
		return fmt.Errorf("error opening QUIC stream: %w", err)
	}

	defer session.Close()

	err = session.mutateState(RpcSessionCreated, RpcSessionOpen)
	if err != nil {
		return fmt.Errorf("error mutating session state: %w", err)
	}

	// read the parameter request for the username

	paramsRequest := &loginParameterRequest{}

	sender, err := readMessageFromUnknown[*loginParameterRequest](session, paramsRequest)
	if err != nil {
		return fmt.Errorf("error reading params request: %w", err)
	}

	username := paramsRequest.username

	session.partner = sender

	// check if the user exists

	db := config.DB()

	var failed = false

	user, err := db.User.Query().Where(user.UsernameEQ(username)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			failed = true
		} else {
			return fmt.Errorf("error reading params request: %w", err)
		}
	}

	// return the client hashing parameters, return a decoy if the user does not exist

	var clientHashing util.ArgonParameters
	if failed {
		clientHashing, err = util.GenerateDecoyArgonParametersFromSeed([]byte(username), pki.GetSeed())
		if err != nil {
			return fmt.Errorf("error generating argon parameters: %w", err)
		}
	} else {
		clientHashing = *user.PasswordClientHashingOptions
	}

	loginParams := &loginParameters{
		passwordParams: clientHashing,
	}

	err = WriteMessage[*loginParameters](session, loginParams)
	if err != nil {
		return fmt.Errorf("error writing login parameters: %w", err)
	}

	// read the login request

	login := &loginRequest{}

	err = ReadMessage[*loginRequest](session, login)
	if err != nil {
		return fmt.Errorf("error reading login request: %w", err)
	}

	// check the password hash
	err = util.VerifyPassword(login.passwordHash, user.PasswordDoubleHashed, *user.PasswordServerHashingOptions)
	if err != nil {
		return fmt.Errorf("error verifying password: %w", err)
	}

	// check the totp code
	if !util.ValidateTotp(user.TotpSecret, login.totp) {
		return fmt.Errorf("error validating totp: %w", err)
	}

	// login successful, return the certificate and encrypted private key

	success := &loginSuccessResponse{
		Cert:                user.Certificate,
		EncryptedPrivateKey: user.EncryptedPrivateKey,
	}

	err = WriteMessage[*loginSuccessResponse](session, success)
	if err != nil {
		return fmt.Errorf("error writing login success response: %w", err)
	}

	session.Close()

	conn.Close(200, "")

	return nil
}
