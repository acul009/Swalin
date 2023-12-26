package system

import (
	"github.com/rahn-it/svalin/rpc"
	"github.com/rahn-it/svalin/util"
)

type getDevicesCommand struct {
	*SyncDownCommand[string, *DeviceInfo]
}

func CreateGetDevicesCommandHandler(m util.ObservableMap[string, *DeviceInfo]) rpc.RpcCommandHandler {
	return func() rpc.RpcCommand {
		syncCmd := NewSyncDownCommand[string, *DeviceInfo](nil)
		syncCmd.SetSourceMap(m)
		return &getDevicesCommand{
			SyncDownCommand: syncCmd,
		}
	}
}

func NewGetDevicesCommand(targetMap util.UpdateableMap[string, *DeviceInfo]) *getDevicesCommand {
	return &getDevicesCommand{
		SyncDownCommand: NewSyncDownCommand[string, *DeviceInfo](targetMap),
	}
}

func (c *getDevicesCommand) GetKey() string {
	return "get-devices"
}
