package system

import (
	"fmt"
	"log"

	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/rpc"
)

func CreateEnrollDeviceCommandHandler(enrollmentManager rpc.EnrollmentManager) rpc.RpcCommandHandler {
	return func() rpc.RpcCommand {
		return &enrollDeviceCommand{
			enrollmentManager: enrollmentManager,
		}
	}
}

type enrollDeviceCommand struct {
	Cert              *pki.Certificate
	enrollmentManager rpc.EnrollmentManager
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
	err := c.enrollmentManager.AcceptEnrollment(c.Cert)
	if err != nil {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 500,
			Msg:  "Internal Server Error",
		})
		return fmt.Errorf("error accepting enrollment: %w", err)
	}

	err = session.WriteResponseHeader(rpc.SessionResponseHeader{
		Code: 200,
		Msg:  "OK",
	})

	return nil
}

func (c *enrollDeviceCommand) ExecuteClient(session *rpc.RpcSession) error {
	log.Printf("sending enrollment successful")
	return nil
}
