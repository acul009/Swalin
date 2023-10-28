package rpc

import (
	"context"
	"fmt"
	"log"
	"rahnit-rmm/config"
	"rahnit-rmm/ent"
	"rahnit-rmm/ent/user"
	"rahnit-rmm/pki"
	"rahnit-rmm/util"
)

func Login(conn *RpcConnection, username string, password []byte, totpCode string) error {
	defer conn.Close(500, "")

	credentials, err := pki.GenerateCredentials()
	if err != nil {
		return fmt.Errorf("error generating temp credentials: %w", err)
	}

	conn.credentials = credentials

	session, err := conn.AcceptSession(context.Background())
	if err != nil {
		return fmt.Errorf("error opening QUIC stream: %w", err)
	}

	defer session.Close()

	err = session.mutateState(RpcSessionCreated, RpcSessionOpen)
	if err != nil {
		return fmt.Errorf("error mutating session state: %w", err)
	}

	err = receivePartnerKey(session)
	if err != nil {
		return fmt.Errorf("error receiving partner key: %w", err)
	}

	paramRequest := &loginParameterRequest{
		Username: username,
	}

	err = WriteMessage[*loginParameterRequest](session, paramRequest)
	if err != nil {
		return fmt.Errorf("error writing params request: %w", err)
	}

	params := loginParameters{}

	err = ReadMessage[*loginParameters](session, &params)
	if err != nil {
		return fmt.Errorf("error reading params request: %w", err)
	}

	hash, err := util.HashPassword(password, params.PasswordParams)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	login := &loginRequest{
		PasswordHash: hash,
		Totp:         totpCode,
	}

	err = WriteMessage[*loginRequest](session, login)
	if err != nil {
		return fmt.Errorf("error writing login request: %w", err)
	}

	success := loginSuccessResponse{}

	err = ReadMessage[*loginSuccessResponse](session, &success)
	if err != nil {
		return fmt.Errorf("error reading login response: %w", err)
	}

	privateKey, err := pki.PrivateKeyFromBinary(success.EncryptedPrivateKey, password)
	if err != nil {
		return fmt.Errorf("error decrypting private key: %w", err)
	}

	err = pki.SaveUserCredentials(username, password, success.Cert, privateKey)
	if err != nil {
		return fmt.Errorf("error saving user credentials: %w", err)
	}

	err = pki.Root.Set(success.RootCert)
	if err != nil {
		return fmt.Errorf("error saving root cert: %w", err)
	}

	err = pki.Upstream.Set(success.UpstreamCert)
	if err != nil {
		return fmt.Errorf("error saving upstream cert: %w", err)
	}

	config.V().Set("upstream.address", conn.connection.RemoteAddr().String())
	err = config.Save()
	if err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	err = session.Close()
	if err != nil {
		return fmt.Errorf("error closing session: %w", err)
	}

	conn.Close(200, "done")

	return nil
}

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

func acceptLoginRequest(conn *RpcConnection) error {
	// prepare session
	ctx := conn.connection.Context()

	var err error

	defer func() {
		if err != nil {
			conn.Close(500, "")
		}
	}()

	log.Printf("Incoming login request...")

	session, err := conn.OpenSession(ctx)
	if err != nil {
		return fmt.Errorf("error opening QUIC stream: %w", err)
	}

	defer session.Close()

	err = session.mutateState(RpcSessionCreated, RpcSessionOpen)
	if err != nil {
		return fmt.Errorf("error mutating session state: %w", err)
	}

	log.Printf("Session opened, sending public key")

	err = sendMyKey(session)
	if err != nil {
		return fmt.Errorf("error sending public key: %w", err)
	}

	// read the parameter request for the username

	log.Printf("reading params request...")

	paramsRequest := loginParameterRequest{}

	sender, err := readMessageFromUnknown[*loginParameterRequest](session, &paramsRequest)
	if err != nil {
		return fmt.Errorf("error reading params request: %w", err)
	}

	username := paramsRequest.Username

	session.partner = sender

	log.Printf("Received params request with username: %s\n", username)

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
		log.Printf("User %s does not exist, generating decoy", username)
		clientHashing, err = util.GenerateDecoyArgonParametersFromSeed([]byte(username), pki.GetSeed())
		if err != nil {
			return fmt.Errorf("error generating argon parameters: %w", err)
		}
	} else {
		log.Printf("User %s exists, using existing parameters %+v", username, user.PasswordClientHashingOptions)
		clientHashing = *user.PasswordClientHashingOptions
	}

	loginParams := loginParameters{
		PasswordParams: clientHashing,
	}

	err = WriteMessage[*loginParameters](session, &loginParams)
	if err != nil {
		return fmt.Errorf("error writing login parameters: %w", err)
	}

	// read the login request

	login := loginRequest{}

	err = ReadMessage[*loginRequest](session, &login)
	if err != nil {
		return fmt.Errorf("error reading login request: %w", err)
	}

	if failed {
		return fmt.Errorf("user does not exist")
	}

	// check the password hash
	err = util.VerifyPassword(login.PasswordHash, user.PasswordDoubleHashed, *user.PasswordServerHashingOptions)
	if err != nil {
		return fmt.Errorf("error verifying password: %w", err)
	}

	// check the totp code
	if !util.ValidateTotp(user.TotpSecret, login.Totp) {
		return fmt.Errorf("error validating totp: %w", err)
	}

	// login successful, return the certificate and encrypted private key
	cert, err := pki.CertificateFromPem([]byte(user.Certificate))
	if err != nil {
		return fmt.Errorf("error parsing user certificate: %w", err)
	}

	rootCert, err := pki.Root.Get()
	if err != nil {
		return fmt.Errorf("error loading root certificate: %w", err)
	}

	hostcredentials, err := pki.GetHostCredentials()
	if err != nil {
		return fmt.Errorf("error loading host credentials: %w", err)
	}

	serverCert, err := hostcredentials.GetCertificate()
	if err != nil {
		return fmt.Errorf("error loading current certificate: %w", err)
	}

	success := &loginSuccessResponse{
		RootCert:            rootCert,
		UpstreamCert:        serverCert,
		Cert:                cert,
		EncryptedPrivateKey: user.EncryptedPrivateKey,
	}

	err = WriteMessage[*loginSuccessResponse](session, success)
	if err != nil {
		return fmt.Errorf("error writing login success response: %w", err)
	}

	session.Close()

	return nil
}
