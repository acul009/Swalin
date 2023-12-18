package rmm

import (
	"github.com/rahn-it/svalin/rpc"
	"github.com/rahn-it/svalin/system"
	"github.com/rahn-it/svalin/util"
)

type getDevicesCommand struct {
	*system.SyncDownCommand[string, *system.DeviceInfo]
}

func CreateGetDevicesCommandHandler(m util.ObservableMap[string, *system.DeviceInfo]) rpc.RpcCommandHandler {
	return func() rpc.RpcCommand {
		syncCmd := system.NewSyncDownCommand[string, *system.DeviceInfo](nil)
		syncCmd.SetSourceMap(m)
		return &getDevicesCommand{
			SyncDownCommand: syncCmd,
		}
	}
}

func NewGetDevicesCommand(targetMap util.UpdateableMap[string, *system.DeviceInfo]) *getDevicesCommand {
	return &getDevicesCommand{
		SyncDownCommand: system.NewSyncDownCommand[string, *system.DeviceInfo](targetMap),
	}
}

func (c *getDevicesCommand) GetKey() string {
	return "get-devices"
}
