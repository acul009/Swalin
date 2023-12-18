package rmm

import (
	"fmt"

	"github.com/rahn-it/svalin/rpc"
	"github.com/rahn-it/svalin/system"
	"github.com/rahn-it/svalin/util"
)

func MonitorProcessesCommandHandler() rpc.RpcCommand {
	return NewMonitorProcessesCommand(nil)
}

type monitorProcessesCommand struct {
	*system.SyncDownCommand[int32, *ProcessInfo]
}

func NewMonitorProcessesCommand(targetMap util.UpdateableMap[int32, *ProcessInfo]) *monitorProcessesCommand {
	return &monitorProcessesCommand{
		SyncDownCommand: system.NewSyncDownCommand[int32, *ProcessInfo](targetMap),
	}
}

func (c *monitorProcessesCommand) GetKey() string {
	return "monitor-processes"
}

func (c *monitorProcessesCommand) ExecuteServer(session *rpc.RpcSession) error {
	errChan := make(chan error)
	processes, err := MonitorProcesses(errChan)
	if err != nil {
		return fmt.Errorf("error monitoring processes: %w", err)
	}

	c.SyncDownCommand.SetSourceMap(processes)

	go func() {
		errChan <- c.SyncDownCommand.ExecuteServer(session)
	}()

	return <-errChan
}
