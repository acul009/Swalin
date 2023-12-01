package rpc

import (
	"fmt"
	"io"
	"log"

	"github.com/rahn-it/svalin/pki"
)

func ForwardCommandHandler() RpcCommand {
	return &forwardCommand{}
}

type forwardCommand struct {
	Target *pki.Certificate
	cmd    RpcCommand
}

func newForwardCommand(target *pki.Certificate, cmd RpcCommand) *forwardCommand {
	return &forwardCommand{
		Target: target,
		cmd:    cmd,
	}
}

func (f *forwardCommand) ExecuteServer(session *RpcSession) error {
	// Verify certificate first

	_, err := session.Verifier().Verify(f.Target)
	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 400,
			Msg:  "Invalid target certificate",
		})
		return fmt.Errorf("invalid target certificate: %w", err)
	}

	// get connection
	conn, err := session.connection.server.getConnectionWith(f.Target)
	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 500,
			Msg:  "Unable to connect to target",
		})
		return fmt.Errorf("unable to connect to target: %w", err)
	}

	// open session
	forwardSession, err := conn.OpenSession(session.Context())
	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 500,
			Msg:  "Unable to open session to target",
		})
		return fmt.Errorf("unable to open session to target: %w", err)
	}

	defer func() {
		err := forwardSession.Close()
		if err != nil {
			panic(err)
		}
	}()

	err = forwardSession.mutateState(RpcSessionCreated, RpcSessionOpen)
	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 500,
			Msg:  "Error mutating session state",
		})
		return fmt.Errorf("error mutating session state: %w", err)
	}

	err = session.WriteResponseHeader(SessionResponseHeader{
		Code: 200,
		Msg:  "OK",
	})

	if err != nil {
		return fmt.Errorf("error writing response header: %w", err)
	}

	errChan := make(chan error)

	go func() {
		_, err := io.Copy(session, forwardSession)
		errChan <- err
	}()
	go func() {
		_, err := io.Copy(forwardSession, session)
		errChan <- err
	}()

	err = <-errChan

	log.Printf("Forwarding session closed: %v", err)

	if err != nil {
		return fmt.Errorf("error forwarding session: %w", err)
	}

	return nil
}

func (f *forwardCommand) ExecuteClient(session *RpcSession) error {

	err := session.mutateState(RpcSessionOpen, RpcSessionCreated)
	if err != nil {
		return fmt.Errorf("error mutating session state: %w", err)
	}

	session.partnerKey = f.Target.PublicKey()

	log.Printf("Session forwarded, starting next command...")

	running, err := session.sendCommand(f.cmd)
	if err != nil {
		return fmt.Errorf("error sending forwarded command: %w", err)
	}

	err = running.Wait()
	if err != nil {
		return fmt.Errorf("error executing forwarded command: %w", err)
	}

	return nil
}

func (f *forwardCommand) GetKey() string {
	return "forward"
}
