package rpc

import (
	"context"
	"fmt"
	"log"

	"github.com/rahn-it/svalin/pki"
)

func Login(conn *RpcConnection, loginHandler func(*RpcSession) error) error {
	defer conn.Close(500, "")

	tempCredentials, err := pki.GenerateCredentials()
	if err != nil {
		return fmt.Errorf("error generating temp credentials: %w", err)
	}

	conn.credentials = tempCredentials
	session, err := conn.OpenSession(context.Background())
	if err != nil {
		return fmt.Errorf("error opening QUIC stream: %w", err)
	}

	defer session.Close()

	err = session.mutateState(RpcSessionCreated, RpcSessionOpen)
	if err != nil {
		return fmt.Errorf("error mutating session state: %w", err)
	}

	err = exchangeKeys(session)
	if err != nil {
		return fmt.Errorf("error exchanging keys: %w", err)
	}

	err = loginHandler(session)
	if err != nil {
		return fmt.Errorf("error executing login handler: %w", err)
	}

	return nil
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

	session, err := conn.AcceptSession(ctx)
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
