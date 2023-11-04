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
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/fyne-io/terminal"
)

type deviceView struct {
	ep        *rpc.RpcEndpoint
	device    rpc.DeviceInfo
	ctx       context.Context
	cancel    context.CancelFunc
	container *fyne.Container
	osBind    binding.String
	cpuBind   binding.String
	memBind   binding.String
	active    util.UpdateableObservable[*rmm.ActiveStats]
	static    util.UpdateableObservable[*rmm.StaticStats]
}

func newDeviceView(ep *rpc.RpcEndpoint, device rpc.DeviceInfo) *deviceView {
	osBind := binding.NewString()
	memBind := binding.NewString()

	d := &deviceView{
		ep:      ep,
		device:  device,
		osBind:  osBind,
		memBind: memBind,
		active:  util.NewObservable[*rmm.ActiveStats](nil),
		static:  util.NewObservable[*rmm.StaticStats](nil),
	}

	cpuDisplay := newCpuDisplay(util.DeriveObservable[*rmm.ActiveStats, *rmm.CpuStats](d.active, func(active *rmm.ActiveStats) *rmm.CpuStats {
		if active == nil {
			return nil
		}
		return active.Cpu
	}))

	d.container = container.NewVBox(
		widget.NewLabel(device.Name()),
		widget.NewLabelWithData(osBind),
		container.NewHBox(
			cpuDisplay,
			widget.NewLabelWithData(memBind),
		),
		widget.NewButton("Terminal", func() {
			// create pipe fifo
			readInput, writeInput := io.Pipe()
			readOutput, writeOutput := io.Pipe()

			ctx, cancel := context.WithCancel(context.Background())

			remShell := rmm.NewRemoteShellCommand(readInput, writeOutput)
			go func() {
				err := d.ep.SendCommandTo(ctx, device.Certificate, remShell)
				if err != nil {
					panic(err)
				}
			}()

			term := terminal.New()
			go func() {
				err := term.RunWithConnection(writeInput, readOutput)
				term.RunLocalShell()
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
	)

	return d
}

func (d *deviceView) Prepare() fyne.CanvasObject {
	d.ctx, d.cancel = context.WithCancel(context.Background())

	go func() {

		cmd := rmm.NewMonitorSystemCommand(d.static, d.active)
		err := d.ep.SendCommandTo(d.ctx, d.device.Certificate, cmd)
		if err != nil {
			panic(err)
		}
	}()

	return d.container
}

func (d *deviceView) Close() {
	log.Printf("closing device view...")
	d.cancel()
}
