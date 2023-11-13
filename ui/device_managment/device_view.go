package managment

import (
	"context"
	"io"
	"log"
	"rahnit-rmm/rmm"
	"rahnit-rmm/rpc"
	"rahnit-rmm/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/fyne-io/terminal"
)

type deviceView struct {
	ep                *rpc.RpcEndpoint
	device            rpc.DeviceInfo
	ctx               context.Context
	cancel            context.CancelFunc
	canvasObject      fyne.CanvasObject
	active            util.UpdateableObservable[*rmm.ActiveStats]
	services          util.UpdateableObservable[*rmm.ServiceStats]
	static            util.UpdateableObservable[*rmm.StaticStats]
	runningMonitoring util.AsyncAction
	runningServices   util.AsyncAction
	tunnelDisp        *tunnelDisplay
}

func newDeviceView(ep *rpc.RpcEndpoint, device rpc.DeviceInfo) *deviceView {

	d := &deviceView{
		ep:       ep,
		device:   device,
		active:   util.NewObservable[*rmm.ActiveStats](nil),
		static:   util.NewObservable[*rmm.StaticStats](nil),
		services: util.NewObservable[*rmm.ServiceStats](nil),
	}

	cpuDisplay := newCpuDisplay(util.DeriveObservable[*rmm.ActiveStats, *rmm.CpuStats](d.active, func(active *rmm.ActiveStats) *rmm.CpuStats {
		if active == nil {
			return nil
		}
		return active.Cpu
	}))

	processList := newProcessList(util.DeriveObservable[*rmm.ActiveStats, *rmm.ProcessStats](d.active, func(active *rmm.ActiveStats) *rmm.ProcessStats {
		if active == nil {
			return nil
		}
		return active.Processes
	}))

	performance := container.NewVBox(
		widget.NewLabel(device.Name()),
		container.NewHBox(
			cpuDisplay,
		),
		widget.NewButton("Terminal", func() {
			// create pipe fifo
			readInput, writeInput := io.Pipe()
			readOutput, writeOutput := io.Pipe()

			ctx, cancel := context.WithCancel(context.Background())

			remShell := rmm.NewRemoteShellCommand(readInput, writeOutput)
			shellConn, err := d.ep.SendCommandTo(ctx, device.Certificate, remShell)
			if err != nil {
				panic(err)
			}

			go func() {
				err := shellConn.Wait()
				if err != nil {
					panic(err)
				}
			}()

			term := terminal.New()
			go func() {
				err := term.RunWithConnection(writeInput, readOutput)
				if err != nil {
					panic(err)
				}
			}()

			window := fyne.CurrentApp().NewWindow("Terminal")

			window.SetContent(
				term,
			)

			window.SetOnClosed(func() {
				cancel()
			})

			window.Resize(fyne.NewSize(800, 600))

			window.Show()

		}),
		processList,
	)

	serviceList := newServiceList(d.services)

	services := container.NewVBox(
		serviceList,
	)

	d.tunnelDisp = newTunnelDisplay(d.ep, device)

	d.canvasObject = container.NewAppTabs(
		container.NewTabItem("Performance", performance),
		container.NewTabItem("Services", services),
		container.NewTabItem("Tunnels", d.tunnelDisp),
	)

	return d
}

func (d *deviceView) Prepare() fyne.CanvasObject {
	d.ctx, d.cancel = context.WithCancel(context.Background())

	monitorCmd := rmm.NewMonitorSystemCommand(d.static, d.active)
	runningMonitoring, err := d.ep.SendCommandTo(d.ctx, d.device.Certificate, monitorCmd)
	if err != nil {
		panic(err)
	}

	d.runningMonitoring = runningMonitoring

	servicesCmd := rmm.NewMonitorServicesCommand(d.services)
	runningServices, err := d.ep.SendCommandTo(d.ctx, d.device.Certificate, servicesCmd)
	if err != nil {
		panic(err)
	}

	go func() {
		err := runningServices.Wait()
		if err != nil {
			panic(err)
		}
	}()

	d.runningServices = runningServices

	return d.canvasObject
}

func (d *deviceView) Close() {
	log.Printf("closing device view...")
	d.cancel()
	err := d.runningMonitoring.Close()
	if err != nil {
		panic(err)
	}
}
