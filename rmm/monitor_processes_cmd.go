package rmm

import (
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

}
