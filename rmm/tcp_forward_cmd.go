package rmm

import (
	"net"
	"rahnit-rmm/rpc"
)

type tcpForwardCommand struct {
	Target string
	conn   net.Conn
}

func (c *tcpForwardCommand) GetKey() string {
	return "tcp-forward"
}

func (c *tcpForwardCommand) ExecuteServer(session *rpc.RpcSession) error {

}

func (c *tcpForwardCommand) ExecuteClient(session *rpc.RpcSession) error {
	return nil
}
