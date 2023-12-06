package rpc

import (
	"context"
	"fmt"
	"log"

	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/util"
)

func Login(conn *RpcConnection, username string, password []byte, totpCode string) (*EndPointInitInfo, error) {
	defer conn.Close(500, "")

	credentials, err := pki.GenerateCredentials()
	if err != nil {
		return nil, fmt.Errorf("error generating temp credentials: %w", err)
	}

	conn.credentials = credentials

	session, err := conn.AcceptSession(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error opening QUIC stream: %w", err)
	}

	defer session.Close()

	err = session.mutateState(RpcSessionCreated, RpcSessionOpen)
	if err != nil {
		return nil, fmt.Errorf("error mutating session state: %w", err)
	}

	err = exchangeKeys(session)
	if err != nil {
		return nil, fmt.Errorf("error exchanging keys: %w", err)
	}

	paramRequest := &loginParameterRequest{
		Username: username,
	}

	err = WriteMessage[*loginParameterRequest](session, paramRequest)
	if err != nil {
		return nil, fmt.Errorf("error writing params request: %w", err)
	}

	params := loginParameters{}

	err = ReadMessage[*loginParameters](session, &params)
	if err != nil {
		return nil, fmt.Errorf("error reading params request: %w", err)
	}

	hash, err := util.HashPassword(password, params.PasswordParams)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	loginReq := &loginRequest{
		PasswordHash: hash,
		Totp:         totpCode,
	}

	err = WriteMessage[*loginRequest](session, loginReq)
	if err != nil {
		return nil, fmt.Errorf("error writing login request: %w", err)
	}

	success := loginSuccessResponse{}

	err = ReadMessage[*loginSuccessResponse](session, &success)
	if err != nil {
		return nil, fmt.Errorf("error reading login response: %w", err)
	}

	privateKey, err := pki.PrivateKeyFromBinary(success.EncryptedPrivateKey, password)
	if err != nil {
		return nil, fmt.Errorf("error decrypting private key: %w", err)
	}

	login := &EndPointInitInfo{
		Root:        success.RootCert,
		Upstream:    success.UpstreamCert,
		Credentials: pki.CredentialsFromCertAndKey(success.Cert, privateKey),
	}

	session.Close()

	conn.Close(200, "done")

	return login, nil
}

func (s *RpcServer) acceptLoginRequest(conn *RpcConnection) error {
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

	err = exchangeKeys(session)
	if err != nil {
		return fmt.Errorf("error exchanging keys: %w", err)
	}

	err = s.loginHandler(session)
	session.Close()
	if err != nil {
		conn.Close(500, "error during login")
		return fmt.Errorf("error during login: %w", err)
	}

	conn.Close(200, "done")

	return nil

}
