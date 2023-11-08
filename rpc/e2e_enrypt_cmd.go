package rpc

import (
	"crypto/ecdh"
	"crypto/rand"
	"fmt"
	"log"
	"rahnit-rmm/util"
)

func CreateE2eDecryptCommandHandler(commands *CommandCollection) func() RpcCommand {
	return func() RpcCommand {
		return &e2eEncryptCommand{
			commands: commands,
		}
	}
}

type e2eEncryptCommand struct {
	ClientPublicKey  []byte
	clientPrivateKey *ecdh.PrivateKey
	cmd              RpcCommand
	commands         *CommandCollection
}

type e2eResponse struct {
	ServerPublicKey []byte
	HashingParams   util.ArgonParameters
}

func newE2eEncryptCommand(cmd RpcCommand) (*e2eEncryptCommand, error) {
	curve := ecdh.P521()

	key, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("error generating key: %w", err)
	}

	return &e2eEncryptCommand{
		ClientPublicKey:  key.PublicKey().Bytes(),
		clientPrivateKey: key,
		cmd:              cmd,
	}, nil
}

func (e *e2eEncryptCommand) ExecuteServer(session *RpcSession) error {
	curve := ecdh.P521()

	log.Printf("Encryption requested...")

	remotePub, err := curve.NewPublicKey(e.ClientPublicKey)
	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 400,
			Msg:  "Error parsing public key",
		})
		return fmt.Errorf("error parsing public key: %w", err)
	}

	key, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 500,
			Msg:  "Error generating key",
		})
		return fmt.Errorf("error generating key: %w", err)
	}

	log.Printf("Key generated")

	shared, err := key.ECDH(remotePub)
	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 500,
			Msg:  "Error computing shared secret",
		})
		return fmt.Errorf("error computing shared secret: %w", err)
	}

	log.Printf("Shared secret computed")

	argonParams, err := util.GenerateArgonParameters(util.ArgonStrengthDefault)
	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 500,
			Msg:  "Error generating argon parameters",
		})
		return fmt.Errorf("error generating argon parameters: %w", err)
	}

	encryptionKey, err := util.HashPassword(shared, argonParams)
	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 500,
			Msg:  "Error deriving encryption key",
		})
		return fmt.Errorf("error deriving encryption key: %w", err)
	}

	log.Printf("initializing crypto stream...")

	cryptoStream, err := util.NewDefaultCipherStream(session.stream, encryptionKey)
	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 500,
			Msg:  "Error creating crypto stream",
		})
		return fmt.Errorf("error creating crypto stream: %w", err)
	}

	log.Printf("writing ok response...")

	err = session.WriteResponseHeader(SessionResponseHeader{
		Code: 200,
		Msg:  "OK",
	})
	if err != nil {
		return fmt.Errorf("error writing response header: %w", err)
	}

	log.Printf("writing own public key and iv...")

	err = WriteMessage[e2eResponse](session, e2eResponse{
		ServerPublicKey: key.PublicKey().Bytes(),
		HashingParams:   argonParams,
	})
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}

	log.Printf("starting encryption...")

	session.stream = cryptoStream

	err = session.mutateState(RpcSessionOpen, RpcSessionCreated)
	if err != nil {
		return fmt.Errorf("error mutating session state: %w", err)
	}

	err = session.handleIncoming(e.commands)
	if err != nil {
		return fmt.Errorf("error handling encrypted session: %w", err)
	}

	return nil
}

func (e *e2eEncryptCommand) ExecuteClient(session *RpcSession) error {

	fmt.Printf("Trying to encrypt session...\n")

	resp := &e2eResponse{}
	err := ReadMessage[*e2eResponse](session, resp)
	if err != nil {
		return fmt.Errorf("error reading message: %w", err)
	}

	curve := ecdh.P521()
	pub, err := curve.NewPublicKey(resp.ServerPublicKey)
	if err != nil {
		return fmt.Errorf("error parsing public key: %w", err)
	}

	shared, err := e.clientPrivateKey.ECDH(pub)
	if err != nil {
		return fmt.Errorf("error computing shared secret: %w", err)
	}

	encryptionKey, err := util.HashPassword(shared, resp.HashingParams)
	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 500,
			Msg:  "Error deriving encryption key",
		})
		return fmt.Errorf("error deriving encryption key: %w", err)
	}

	cipherStream, err := util.NewDefaultCipherStream(session.stream, encryptionKey)
	if err != nil {
		return fmt.Errorf("error creating crypto stream: %w", err)
	}

	session.stream = cipherStream

	err = session.mutateState(RpcSessionOpen, RpcSessionCreated)
	if err != nil {
		return fmt.Errorf("error mutating session state: %w", err)
	}

	log.Printf("Session encrypted, sending command...")

	running, err := session.sendCommand(e.cmd)
	if err != nil {
		return fmt.Errorf("error sending encrypted command: %w", err)
	}

	err = running.Wait()
	if err != nil {
		return fmt.Errorf("error executing encrypted command: %w", err)
	}

	return nil
}

func (e *e2eEncryptCommand) GetKey() string {
	return "e2e-encrypt"
}
