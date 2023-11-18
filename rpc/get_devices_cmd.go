package rpc

import (
	"rahnit-rmm/util"
)

type getDevicesCommand struct {
	*syncDownCommand[string, *DeviceInfo]
}

func GetDevicesCommandHandler() RpcCommand {
	return &getDevicesCommand{
		syncDownCommand: NewSyncDownCommand[string, *DeviceInfo](nil),
	}
}

func NewGetDevicesCommand(targetMap util.ObservableMap[string, *DeviceInfo]) *getDevicesCommand {
	return &getDevicesCommand{
		syncDownCommand: NewSyncDownCommand[string, *DeviceInfo](targetMap),
	}
}

func (c *getDevicesCommand) GetKey() string {
	return "get-devices"
}

func (c *getDevicesCommand) ExecuteServer(session *RpcSession) error {
	devicemap := session.connection.server.devices.devices
	c.syncDownCommand.targetMap = devicemap
	return c.syncDownCommand.ExecuteServer(session)
}
