package rpc

import (
	"fmt"
	"rahnit-rmm/config"
	"rahnit-rmm/ent/user"
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
	Cert                []byte
	EncryptedPrivateKey []byte
}

func acceptLoginRequest(conn *rpcConnection) error {
	ctx := conn.connection.Context()

	session, err := conn.OpenSession(ctx)
	if err != nil {
		return fmt.Errorf("error opening QUIC stream: %w", err)
	}

	err = session.mutateState(RpcSessionCreated, RpcSessionOpen)
	if err != nil {
		return fmt.Errorf("error mutating session state: %w", err)
	}
	paramsRequest := &loginParameterRequest{}

	sender, err := readMessageFromUnknown[*loginParameterRequest](session, paramsRequest)
	if err != nil {
		return fmt.Errorf("error reading params request: %w", err)
	}

	db := config.DB()

	var failed = false

	user, err := db.User.Query().Where(user.UsernameEQ(paramsRequest.username)).Only(ctx)
	if err != nil {
		if ... {
			failed = true
		} else {
			return fmt.Errorf("error reading params request: %w", err)
		}
	}
}
