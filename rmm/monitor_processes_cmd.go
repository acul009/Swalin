package rmm

import (
	"fmt"
	"rahnit-rmm/rpc"
	"rahnit-rmm/util"
)

type monitorProcessesCommand struct {
	*rpc.SyncDownCommand[int32, *ProcessInfo]
}

func NewMonitorProcessesCommand(targetMap util.ObservableMap[int32, *ProcessInfo]) *monitorProcessesCommand {
	return &monitorProcessesCommand{
		SyncDownCommand: rpc.NewSyncDownCommand[int32, *ProcessInfo](targetMap),
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

	c.SyncDownCommand.SetMap(processes)

	go func() {
		errChan <- c.SyncDownCommand.ExecuteServer(session)
	}()

	return <-errChan
}
