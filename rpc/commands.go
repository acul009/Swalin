package rpc

type RpcCommandHandler[T any] func() RpcCommand[T]

type RpcCommand[T any] interface {
	GetKey() string
	ExecuteServer(session *RpcSession[T]) error
	ExecuteClient(session *RpcSession[T]) error
}

type CommandCollection[T any] struct {
	Commands map[string]RpcCommandHandler[T]
}

func NewCommandCollection[T any](commands ...RpcCommandHandler[T]) *CommandCollection[T] {
	collection := &CommandCollection[T]{
		Commands: make(map[string]RpcCommandHandler[T]),
	}
	for _, cmd := range commands {
		collection.Add(cmd)
	}

	return collection
}

func (c *CommandCollection[T]) Add(cmdHandler RpcCommandHandler[T]) {
	c.Commands[cmdHandler().GetKey()] = cmdHandler
}

func (c *CommandCollection[T]) Get(cmd string) (RpcCommandHandler[T], bool) {
	commandHandler, ok := c.Commands[cmd]
	return commandHandler, ok
}
