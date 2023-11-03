package rmm

import (
	"fmt"
	"log"
	"rahnit-rmm/rpc"
	"rahnit-rmm/util"
	"time"
)

const reportingInterval = 1 * time.Second

func MonitorSystemCommandHandler() rpc.RpcCommand {
	return &monitorSystemCommand{}
}

type monitorSystemCommand struct {
	static util.Observable[*StaticStats]
	active util.Observable[*ActiveStats]
}

func NewMonitorSystemCommand(static func(*StaticStats), active func(*ActiveStats)) *monitorSystemCommand {
	staticOb := util.NewGenericObservable[*StaticStats](nil)
	staticOb.Subscribe(static)
	activeOb := util.NewGenericObservable[*ActiveStats](nil)
	activeOb.Subscribe(active)
	return &monitorSystemCommand{
		static: staticOb,
		active: activeOb,
	}
}

func (cmd *monitorSystemCommand) GetKey() string {
	return "monitor-system"
}

func (cmd *monitorSystemCommand) ExecuteServer(session *rpc.RpcSession) error {

	static, err := GetStaticStats()
	if err != nil {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 500,
			Msg:  "Unable to read host info",
		})
		return fmt.Errorf("error reading host info: %w", err)
	}

	err = session.WriteResponseHeader(rpc.SessionResponseHeader{
		Code: 200,
		Msg:  "OK",
	})
	if err != nil {
		return fmt.Errorf("error writing response header: %w", err)
	}

	err = rpc.WriteMessage[*StaticStats](session, static)
	if err != nil {
		return fmt.Errorf("error writing static stats: %w", err)
	}

	for {
		active, err := GetActiveStats()
		if err != nil {
			return fmt.Errorf("error getting active stats: %w", err)
		}

		log.Printf("Active stats: %+v\n", active)

		err = rpc.WriteMessage[*ActiveStats](session, active)
		if err != nil {
			return fmt.Errorf("error writing active stats: %w", err)
		}

		time.Sleep(reportingInterval)
	}
}

func (cmd *monitorSystemCommand) ExecuteClient(session *rpc.RpcSession) error {
	log.Printf("Monitoring remote system...")

	static := &StaticStats{}
	err := rpc.ReadMessage[*StaticStats](session, static)
	if err != nil {
		return fmt.Errorf("error reading static stats: %w", err)
	}

	cmd.static.Set(static)

	active := &ActiveStats{}

	for {
		err = rpc.ReadMessage[*ActiveStats](session, active)
		if err != nil {
			return fmt.Errorf("error reading active stats: %w", err)
		}

		cmd.active.Set(active)
	}

	return nil
}
