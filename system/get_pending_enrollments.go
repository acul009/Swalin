package system

import (
	"github.com/rahn-it/svalin/rpc"
	"github.com/rahn-it/svalin/util"
)

var _ rpc.RpcCommand = (*getPendingEnrollmentsCommand)(nil)

func CreateGetEnrollmentsCommandHandler(sourceMap util.ObservableMap[string, *rpc.Enrollment]) rpc.RpcCommandHandler {
	return func() rpc.RpcCommand {
		cmd := NewGetPendingEnrollmentsCommand(nil)
		cmd.SetSourceMap(sourceMap)
		return cmd
	}
}

type getPendingEnrollmentsCommand struct {
	*SyncDownCommand[string, *rpc.Enrollment]
}

func NewGetPendingEnrollmentsCommand(targetMap util.UpdateableMap[string, *rpc.Enrollment]) *getPendingEnrollmentsCommand {
	return &getPendingEnrollmentsCommand{
		SyncDownCommand: NewSyncDownCommand[string, *rpc.Enrollment](targetMap),
	}
}

func (e *getPendingEnrollmentsCommand) GetKey() string {
	return "get-pending-enrollments"
}
