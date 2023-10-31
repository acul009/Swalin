package rpc

import (
	"fmt"
	"rahnit-rmm/pki"
)

func EnrollAgentHandler() RpcCommand {
	return &enrollAgentCommand{}
}

func NewEnrollAgentCommand(agentCert *pki.Certificate) *enrollAgentCommand {
	return &enrollAgentCommand{
		AgentCert: agentCert,
	}
}

type enrollAgentCommand struct {
	AgentCert *pki.Certificate
}

func (c *enrollAgentCommand) GetKey() string {
	return "enroll-agent"
}

func (c *enrollAgentCommand) ExecuteServer(session *RpcSession) error {
	err := session.connection.server.enrollment.acceptEnrollment(c.AgentCert)
	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 500,
			Msg:  "Internal Server Error",
		})
		return fmt.Errorf("error accepting enrollment: %w", err)
	}

	err = session.connection.server.devices.AddDeviceToDB(c.AgentCert)
	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 500,
			Msg:  "Internal Server Error",
		})
		return fmt.Errorf("error adding device: %w", err)
	}

	err = session.WriteResponseHeader(SessionResponseHeader{
		Code: 200,
		Msg:  "OK",
	})

	if err != nil {
		return fmt.Errorf("error writing response header: %w", err)
	}

	return nil
}

func (c *enrollAgentCommand) ExecuteClient(session *RpcSession) error {
	return nil
}
