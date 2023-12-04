package rmm

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/rahn-it/svalin/system"
	"github.com/rahn-it/svalin/util"
)

type Device struct {
	*system.DeviceInfo
	c            *Client
	mutex        sync.Mutex
	processes    util.UpdateableMap[int32, *ProcessInfo]
	tunnelConfig util.Observable[*TunnelConfig]
}

func (d *Device) Name() string {
	return d.Certificate.GetName()
}

func (d *Device) Processes() util.UpdateableMap[int32, *ProcessInfo] {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.processes == nil {
		var pRunning util.AsyncAction

		d.processes = util.NewSyncedMap[int32, *ProcessInfo](
			func(m util.UpdateableMap[int32, *ProcessInfo]) {
				cmd := NewMonitorProcessesCommand(m)
				running, err := d.c.dispatch().SendCommandTo(context.Background(), d.Certificate, cmd)
				if err != nil {
					log.Printf("error subscribing to processes: %v", err)
					return
				}
				pRunning = running
			},
			func(_ util.UpdateableMap[int32, *ProcessInfo]) {
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

func (d *Device) TunnelConfig() util.Observable[*TunnelConfig] {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.tunnelConfig == nil {
		var cRunning util.AsyncAction
		d.tunnelConfig = util.NewSyncedObservable[*TunnelConfig](
			func(uo util.UpdateableObservable[*TunnelConfig]) {
				cmd := NewGetConfigCommand[*TunnelConfig](d.Certificate, uo)
				running, err := d.c.dispatch().SendCommand(context.Background(), cmd)
				if err != nil {
					log.Printf("error subscribing to tunnel config: %v", err)
					return
				}

				cRunning = running
			},
			func(uo util.UpdateableObservable[*TunnelConfig]) {
				err := cRunning.Close()
				if err != nil {
					log.Printf("error unsubscribing from tunnel config: %v", err)
				}
			},
		)
	}

	return d.tunnelConfig
}
