package system

import (
	"fmt"
	"log"

	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/rpc"
)

func CreateEnrollDeviceCommandHandler(enrollmentManager rpc.EnrollmentManager, verifier pki.Verifier, onSuccess func(cert *pki.Certificate) error) rpc.RpcCommandHandler {
	return func() rpc.RpcCommand {
		return &enrollDeviceCommand{
			enrollmentManager: enrollmentManager,
			verfifier:         verifier,
			onSuccess:         onSuccess,
		}
	}
}

type enrollDeviceCommand struct {
	Cert              *pki.Certificate
	enrollmentManager rpc.EnrollmentManager
	verfifier         pki.Verifier
	onSuccess         func(cert *pki.Certificate) error
}

func NewEnrollDeviceCommand(cert *pki.Certificate) *enrollDeviceCommand {
	return &enrollDeviceCommand{
		Cert: cert,
	}
}

func (c *enrollDeviceCommand) GetKey() string {
	return "enroll-device"
}

func (c *enrollDeviceCommand) ExecuteServer(session *rpc.RpcSession) error {
	_, err := c.verfifier.Verify(c.Cert)
	if err != nil {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 400,
			Msg:  "Invalid certificate",
		})
		return fmt.Errorf("invalid certificate: %w", err)
	}

	err = c.enrollmentManager.AcceptEnrollment(c.Cert)
	if err != nil {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 500,
			Msg:  "Internal Server Error",
		})
		return fmt.Errorf("error accepting enrollment: %w", err)
	}

	err = c.onSuccess(c.Cert)
	if err != nil {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 500,
			Msg:  "Internal Server Error",
		})
		return fmt.Errorf("error on success: %w", err)
	}

	err = session.WriteResponseHeader(rpc.SessionResponseHeader{
		Code: 200,
		Msg:  "OK",
	})

	return err
}

func (c *enrollDeviceCommand) ExecuteClient(session *rpc.RpcSession) error {
	log.Printf("sending enrollment successful")
	return nil
}
