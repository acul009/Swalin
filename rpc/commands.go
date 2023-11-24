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

func NewCommandCollection(commands ...RpcCommandHandler) *CommandCollection {
	collection := &CommandCollection{
		Commands: make(map[string]RpcCommandHandler),
	}
	for _, cmd := range commands {
		collection.Add(cmd)
	}

	return collection
}

func (c *CommandCollection) Add(cmdHandler RpcCommandHandler) {
	c.Commands[cmdHandler().GetKey()] = cmdHandler
}

func (c *CommandCollection) Get(cmd string) (RpcCommandHandler, bool) {
	commandHandler, ok := c.Commands[cmd]
	return commandHandler, ok
}
