package managment

import (
	"context"
	"fmt"
	"io"
	"log"
	"rahnit-rmm/rmm"
	"rahnit-rmm/rpc"
	"strconv"
	"strings"

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
}

func newDeviceView(ep *rpc.RpcEndpoint, device rpc.DeviceInfo) *deviceView {
	osBind := binding.NewString()
	cpuBind := binding.NewString()
	memBind := binding.NewString()

	d := &deviceView{
		ep:      ep,
		device:  device,
		osBind:  osBind,
		cpuBind: cpuBind,
		memBind: memBind,
	}

	d.container = container.NewVBox(
		widget.NewLabel(device.Name()),
		widget.NewLabelWithData(osBind),
		container.NewHBox(
			widget.NewLabelWithData(cpuBind),
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
		cmd := rmm.NewMonitorSystemCommand(d.setStatic, d.setActive)
		err := d.ep.SendCommandTo(d.ctx, d.device.Certificate, cmd)
		if err != nil {
			panic(err)
		}
	}()

	return d.container
}

func (d *deviceView) setStatic(static *rmm.StaticStats) {
	log.Printf("Static stats: %+v\n", static)
	d.osBind.Set(static.HostInfo.OS)
}

func (d *deviceView) setActive(active *rmm.ActiveStats) {
	percent := 0.0

	sb := &strings.Builder{}

	for _, cpu := range active.CpuUsage {
		sb.WriteString(fmt.Sprintf("%s %%     ", strconv.FormatFloat(cpu, 'f', 0, 64)))
		percent += cpu
	}

	percent /= float64(len(active.CpuUsage))

	//display cpu in percent
	d.cpuBind.Set(sb.String())
}

func (d *deviceView) Close() {
	d.cancel()
}
