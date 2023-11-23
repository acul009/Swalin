package rmm

import (
	"fmt"
	"io"
	"net"
	"rahnit-rmm/rpc"
)

func TcpForwardCommandHandler() rpc.RpcCommand {
	return &tcpForwardCommand{}
}

type tcpForwardCommand struct {
	Target string
	conn   net.Conn
}

func NewTcpForwardCommand(target string, conn net.Conn) *tcpForwardCommand {
	return &tcpForwardCommand{
		Target: target,
		conn:   conn,
	}
}

func (c *tcpForwardCommand) GetKey() string {
	return "tcp-forward"
}

func (c *tcpForwardCommand) ExecuteServer(session *rpc.RpcSession) error {
	conn, err := net.Dial("tcp", c.Target)
	if err != nil {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 500,
			Msg:  "Unable to connect to given target",
		})
	}

	session.WriteResponseHeader(rpc.SessionResponseHeader{
		Code: 200,
		Msg:  "OK",
	})

	errChan := make(chan error)

	go func() {
		_, err := io.Copy(conn, c.conn)
		errChan <- err
	}()

	go func() {
		_, err := io.Copy(c.conn, conn)
		errChan <- err
	}()

	err = <-errChan
	if err != nil {
		return fmt.Errorf("error copying: %w", err)
	}

	return nil
}

func (c *tcpForwardCommand) ExecuteClient(session *rpc.RpcSession) error {
	errChan := make(chan error)

	go func() {
		_, err := io.Copy(session, c.conn)
		errChan <- err
	}()

	go func() {
		_, err := io.Copy(c.conn, session)
		errChan <- err
	}()

	err := <-errChan
	if err != nil {
		return fmt.Errorf("error copying: %w", err)
	}

	return nil
}
