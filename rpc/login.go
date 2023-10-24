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

	"github.com/quic-go/quic-go"
)

func Login(addr string, username string, password []byte, totpCode string) error {
	err := pki.UnlockWithTempKeys()
	if err != nil {
		return fmt.Errorf("error unlocking with temp keys: %w", err)
	}

	tlsConf := GetTlsClientConfig(ProtoClientLogin)

	quicConf := &quic.Config{}

	quicConn, err := quic.DialAddr(context.Background(), addr, tlsConf, quicConf)
	if err != nil {
		qErr, ok := err.(*quic.TransportError)
		if ok && uint8(qErr.ErrorCode) == 120 {
			return fmt.Errorf("server not ready for login: %w", err)
		}
		return fmt.Errorf("error creating QUIC connection: %w", err)
	}

	initNonceStorage = NewNonceStorage()

	conn := newRpcConnection(quicConn, nil, RpcRoleInit, initNonceStorage, nil, ProtoServerInit)

	defer conn.Close(500, "")

	log.Printf("Connection opened to %s\n", addr)

	session, err := conn.OpenSession(context.Background())
	if err != nil {
		return fmt.Errorf("error opening QUIC stream: %w", err)
	}

	defer session.Close()

	err = session.mutateState(RpcSessionCreated, RpcSessionOpen)
	if err != nil {
		return fmt.Errorf("error mutating session state: %w", err)
	}

	paramRequest := &loginParameterRequest{
		username: username,
	}

	err = WriteMessage[*loginParameterRequest](session, paramRequest)
	if err != nil {
		return fmt.Errorf("error writing params request: %w", err)
	}

	params := &loginParameters{}

	serverPub, err := readMessageFromUnknown[*loginParameters](session, params)
	if err != nil {
		return fmt.Errorf("error reading params request: %w", err)
	}

	session.partner = serverPub

	hash, err := util.HashPassword(password, params.passwordParams)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	login := &loginRequest{
		passwordHash: hash,
		totp:         totpCode,
	}

	err = WriteMessage[*loginRequest](session, login)
	if err != nil {
		return fmt.Errorf("error writing login request: %w", err)
	}

	success := &loginSuccessResponse{}

	err = ReadMessage[*loginSuccessResponse](session, success)
	if err != nil {
		return fmt.Errorf("error reading login response: %w", err)
	}

	privateKey, err := pki.PrivateKeyFromBinary(success.EncryptedPrivateKey, password)
	if err != nil {
		return fmt.Errorf("error decrypting private key: %w", err)
	}

	err = pki.SaveCurrentCertAndKey(success.Cert, privateKey, password)
	if err != nil {
		return fmt.Errorf("error saving current cert and key: %w", err)
	}

	err = pki.SaveRootCert(success.RootCert)
	if err != nil {
		return fmt.Errorf("error saving root cert: %w", err)
	}

	err = pki.SaveUpstreamCert(success.UpstreamCert)
	if err != nil {
		return fmt.Errorf("error saving upstream cert: %w", err)
	}

	return nil
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
	RootCert            *pki.Certificate
	UpstreamCert        *pki.Certificate
	Cert                *pki.Certificate
	EncryptedPrivateKey []byte
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
	cert, err := pki.CertificateFromPem([]byte(user.Certificate))
	if err != nil {
		return fmt.Errorf("error parsing user certificate: %w", err)
	}

	rootCert, err := pki.GetRootCert()
	if err != nil {
		return fmt.Errorf("error loading root certificate: %w", err)
	}

	serverCert, err := pki.GetCurrentCert()
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

	conn.Close(200, "")

	return nil
}
