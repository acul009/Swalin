package rpc

type RegisterCmd struct {
	id string
}

func (c *RegisterCmd) ExecuteClient(session *RpcSession) error {

}

func (c *RegisterCmd) ExecuteServer(session *RpcSession) error {

}

func (c *RegisterCmd) GetKey() string {
	return "register"
}
