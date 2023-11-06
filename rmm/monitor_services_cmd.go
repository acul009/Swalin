package rmm

import (
	"fmt"
	"rahnit-rmm/rpc"
	"rahnit-rmm/util"
)

func MonitorServicesCommandHandler() rpc.RpcCommand {
	return &monitorServicesCommand{}
}

type monitorServicesCommand struct {
	services util.UpdateableObservable[[]ServiceInfo]
}

func NewMonitorServicesCommand(services util.UpdateableObservable[[]ServiceInfo]) *monitorServicesCommand {
	return &monitorServicesCommand{
		services: services,
	}
}

func (cmd *monitorServicesCommand) GetKey() string {
	return "manage-services"
}

func (cmd *monitorServicesCommand) ExecuteServer(session *rpc.RpcSession) error {
	system, err := GetServiceSystem()
	if err != nil {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 500,
			Msg:  "Unable to get service system",
		})
		return fmt.Errorf("error getting service system: %w", err)
	}

	session.WriteResponseHeader(rpc.SessionResponseHeader{
		Code: 200,
		Msg:  "OK",
	})

	for {
		services, err := system.ListServices()
		if err != nil {
			return fmt.Errorf("error listing services: %w", err)
		}

		err = rpc.WriteMessage[[]ServiceInfo](session, services)
		if err != nil {
			return fmt.Errorf("error writing services: %w", err)
		}

	}
}

func (cmd *monitorServicesCommand) ExecuteClient(session *rpc.RpcSession) error {

	services := make([]ServiceInfo, 0)

	for {
		err := rpc.ReadMessage[[]ServiceInfo](session, services)
		if err != nil {
			return fmt.Errorf("error reading services: %w", err)
		}

		cmd.services.Update(func(_ []ServiceInfo) []ServiceInfo {
			return services
		})
	}
}
