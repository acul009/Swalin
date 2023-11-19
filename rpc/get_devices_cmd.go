package rpc

import (
	"rahnit-rmm/util"
)

type getDevicesCommand struct {
	*SyncDownCommand[string, *DeviceInfo]
}

func GetDevicesCommandHandler() RpcCommand {
	return &getDevicesCommand{
		SyncDownCommand: NewSyncDownCommand[string, *DeviceInfo](nil),
	}
}

func NewGetDevicesCommand(targetMap util.ObservableMap[string, *DeviceInfo]) *getDevicesCommand {
	return &getDevicesCommand{
		SyncDownCommand: NewSyncDownCommand[string, *DeviceInfo](targetMap),
	}
}

func (c *getDevicesCommand) GetKey() string {
	return "get-devices"
}

func (c *getDevicesCommand) ExecuteServer(session *RpcSession) error {
	devicemap := session.connection.server.devices.devices
	c.SyncDownCommand.targetMap = devicemap
	return c.SyncDownCommand.ExecuteServer(session)
}
