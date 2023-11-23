package rmm

import (
	"fmt"
	"rahnit-rmm/rpc"
)

func KillProcessCommandHandler() rpc.RpcCommand[*Dependencies] {
	return &killProcessCommand{}
}

type killProcessCommand struct {
	Pid int32
}

func NewKillProcessCommand(pid int32) *killProcessCommand {
	return &killProcessCommand{
		Pid: pid,
	}
}

func (c *killProcessCommand) GetKey() string {
	return "kill-process"
}

func (c *killProcessCommand) ExecuteServer(session *rpc.RpcSession[*Dependencies]) error {

	err := KillProcess(c.Pid)
	if err != nil {
		return fmt.Errorf("error killing process: %w", err)
	}

	return nil
}

func (c *killProcessCommand) ExecuteClient(session *rpc.RpcSession[*Dependencies]) error {
	return nil
}
