package rmm

import (
	"context"
	"fmt"
	"log"
	"rahnit-rmm/util"
	"sync"
)

type Device struct {
	*DeviceInfo
	c         *Client
	mutex     sync.Mutex
	processes util.ObservableMap[int32, *ProcessInfo]
}

func (d *Device) Name() string {
	return d.Certificate.GetName()
}

func (d *Device) Processes() util.ObservableMap[int32, *ProcessInfo] {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.processes == nil {
		var pRunning util.AsyncAction

		d.processes = util.NewSyncedMap[int32, *ProcessInfo](
			func(m util.ObservableMap[int32, *ProcessInfo]) {
				cmd := NewMonitorProcessesCommand(m)
				running, err := d.c.dispatch().SendCommandTo(context.Background(), d.Certificate, cmd)
				if err != nil {
					log.Printf("error subscribing to processes: %v", err)
					return
				}
				pRunning = running
			},
			func(_ util.ObservableMap[int32, *ProcessInfo]) {
				err := pRunning.Close()
				if err != nil {
					log.Printf("error unsubscribing from processes: %v", err)
				}
			},
		)
	}
	return d.processes
}

func (d *Device) KillProcess(pid int32) error {
	cmd := NewKillProcessCommand(pid)

	err := d.c.dispatch().SendSyncCommandTo(context.Background(), d.Certificate, cmd)
	if err != nil {
		return fmt.Errorf("error killing process: %w", err)
	}

	return nil
}

func (d *Device) TunnelConfig() *TunnelConfig {

}
