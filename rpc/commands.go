package rpc

import "fmt"

type RpcCommand interface {
	GetKey() string
	ExecuteServer(session *RpcSession) error
	ExecuteClient(session *RpcSession) error
}

type CommandCollection struct {
	Commands map[string]RpcCommand
}

func NewCommandCollection() *CommandCollection {
	return &CommandCollection{
		Commands: make(map[string]RpcCommand),
	}
}

func (c *CommandCollection) Add(cmd RpcCommand) {
	c.Commands[cmd.GetKey()] = cmd
}

func (c *CommandCollection) handleRequest(header SessionRequestHeader, session *RpcSession) error {
	command, ok := c.Commands[header.Cmd]
	if !ok {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 404,
			Msg:  "Command not Found",
		})
		session.Close()
		return fmt.Errorf("Unknown command: %v", header.Cmd)
	}
	session.WriteResponseHeader(SessionResponseHeader{
		Code: 200,
		Msg:  "OK",
	})
	return command.ExecuteServer(session)
}
