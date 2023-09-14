package rpc

type RpcCommandHandler func(request map[string]interface{}) SessionResponseHeader

type CommandCollection struct {
	Commands map[string]RpcCommandHandler
}

func NewCommandCollection() *CommandCollection {
	return &CommandCollection{
		Commands: make(map[string]RpcCommandHandler),
	}
}

func (c *CommandCollection) AddCommand(name string, cmd RpcCommandHandler) {

}
