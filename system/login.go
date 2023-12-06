package system

import (
	"fmt"
	"log"

	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/rpc"
	"github.com/rahn-it/svalin/util"
)

type User struct {
	Certificate          *pki.Certificate
	EncryptedPrivateKey  []byte
	ClientHashingParams  *util.ArgonParameters
	ServerHashingParams  *util.ArgonParameters
	DoubleHashedPassword []byte
	TotpSecret           string
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

type loginRequestHandler struct {
	getUser  func(string) (*User, error)
	seed     []byte
	root     *pki.Certificate
	upstream *pki.Certificate
}

func NewLoginHandler(getUser func(string) (*User, error), seed []byte, root *pki.Certificate, upstream *pki.Certificate) *loginRequestHandler {
	return &loginRequestHandler{
		getUser:  getUser,
		seed:     seed,
		root:     root,
		upstream: upstream,
	}
}

func (h *loginRequestHandler) HandleLoginRequest(session *rpc.RpcSession) error {

	// read the parameter request for the username

	log.Printf("reading params request...")

	paramsRequest := loginParameterRequest{}

	err := rpc.ReadMessage[*loginParameterRequest](session, &paramsRequest)
	if err != nil {
		return fmt.Errorf("error reading params request: %w", err)
	}

	username := paramsRequest.Username

	log.Printf("Received params request with username: %s\n", username)

	// check if the user exists

	failed := false

	user, err := h.getUser(username)
	if err != nil {
		failed = true
		log.Printf("failed to retrieve user for login: %w", err)
	}

	// return the client hashing parameters, return a decoy if the user does not exist

	var clientHashing util.ArgonParameters
	if failed {
		log.Printf("User %s was not found, generating decoy", username)
		clientHashing, err = util.GenerateDecoyArgonParametersFromSeed([]byte(username), h.seed)
		if err != nil {
			return fmt.Errorf("error generating argon parameters: %w", err)
		}
	} else {
		log.Printf("User %s exists, using existing parameters %+v", username, user.ClientHashingParams)
		clientHashing = *user.ClientHashingParams
	}

	loginParams := loginParameters{
		PasswordParams: clientHashing,
	}

	err = rpc.WriteMessage[*loginParameters](session, &loginParams)
	if err != nil {
		return fmt.Errorf("error writing login parameters: %w", err)
	}

	// read the login request

	login := loginRequest{}

	err = rpc.ReadMessage[*loginRequest](session, &login)
	if err != nil {
		return fmt.Errorf("error reading login request: %w", err)
	}

	if failed {
		util.HashPassword(login.PasswordHash, clientHashing)
		return fmt.Errorf("user does not exist")
	}

	// check the password hash
	err = util.VerifyPassword(login.PasswordHash, user.DoubleHashedPassword, *user.ServerHashingParams)
	if err != nil {
		return fmt.Errorf("error verifying password: %w", err)
	}

	// check the totp code
	if !util.ValidateTotp(user.TotpSecret, login.Totp) {
		return fmt.Errorf("error validating totp: %w", err)
	}

	// login successful, return the certificate and encrypted private key

	success := &loginSuccessResponse{
		RootCert:            h.root,
		UpstreamCert:        h.upstream,
		Cert:                user.Certificate,
		EncryptedPrivateKey: user.EncryptedPrivateKey,
	}

	err = rpc.WriteMessage[*loginSuccessResponse](session, success)
	if err != nil {
		return fmt.Errorf("error writing login success response: %w", err)
	}

	session.Close()
	return nil

}

type loginExecutor struct {
	username  string
	password  []byte
	totpCode  string
	onSuccess func(*rpc.EndPointInitInfo) error
}

func NewLoginExecutor(username string, password []byte, totpCode string, onSuccess func(*rpc.EndPointInitInfo) error) *loginExecutor {
	return &loginExecutor{
		username:  username,
		password:  password,
		totpCode:  totpCode,
		onSuccess: onSuccess,
	}
}

func (e *loginExecutor) Login(session *rpc.RpcSession) error {

	paramRequest := &loginParameterRequest{
		Username: e.username,
	}

	err := rpc.WriteMessage[*loginParameterRequest](session, paramRequest)
	if err != nil {
		return fmt.Errorf("error writing params request: %w", err)
	}

	params := loginParameters{}

	err = rpc.ReadMessage[*loginParameters](session, &params)
	if err != nil {
		return fmt.Errorf("error reading params request: %w", err)
	}

	hash, err := util.HashPassword(e.password, params.PasswordParams)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	loginReq := &loginRequest{
		PasswordHash: hash,
		Totp:         e.totpCode,
	}

	err = rpc.WriteMessage[*loginRequest](session, loginReq)
	if err != nil {
		return fmt.Errorf("error writing login request: %w", err)
	}

	success := loginSuccessResponse{}

	err = rpc.ReadMessage[*loginSuccessResponse](session, &success)
	if err != nil {
		return fmt.Errorf("error reading login response: %w", err)
	}

	privateKey, err := pki.PrivateKeyFromBinary(success.EncryptedPrivateKey, e.password)
	if err != nil {
		return fmt.Errorf("error decrypting private key: %w", err)
	}

	credentials := pki.CredentialsFromCertAndKey(success.Cert, privateKey)

	initInfo := rpc.EndPointInitInfo{
		Root:        success.RootCert,
		Upstream:    success.UpstreamCert,
		Credentials: credentials,
	}

	err = e.onSuccess(&initInfo)
	if err != nil {
		return fmt.Errorf("error executing on success: %w", err)
	}

	return nil
}
