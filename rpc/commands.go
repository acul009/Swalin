package rpc

type RpcCommandHandler func() RpcCommand

type RpcCommand interface {
	GetKey() string
	ExecuteServer(session *RpcSession) error
	ExecuteClient(session *RpcSession) error
}

type CommandCollection struct {
	Commands map[string]RpcCommandHandler
}

func NewCommandCollection() *CommandCollection {
	return &CommandCollection{
		Commands: make(map[string]RpcCommandHandler),
	}
}

func (c *CommandCollection) Add(cmdHandler RpcCommandHandler) {
	c.Commands[cmdHandler().GetKey()] = cmdHandler
}

func (c *CommandCollection) Get(cmd string) (RpcCommandHandler, bool) {
	commandHandler, ok := c.Commands[cmd]
	return commandHandler, ok
}
