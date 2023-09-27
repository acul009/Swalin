package rpc

import "fmt"

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

func (c *CommandCollection) handleRequest(header SessionRequestHeader, session *RpcSession) error {
	commandHandler, ok := c.Commands[header.Cmd]
	if !ok {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 404,
			Msg:  "Command not Found",
		})
		session.Close()
		return fmt.Errorf("unknown command: %v", header.Cmd)
	}
	command := commandHandler()
	reEncode(&header.Args, &command)
	session.ReadyToWrite = false
	return command.ExecuteServer(session)
}
