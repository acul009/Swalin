package managment

import (
	"log"
	"rahnit-rmm/rpc"
	"rahnit-rmm/ui/mainview.go"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type deviceList struct {
	main           *mainview.MainView
	ep             *rpc.RpcEndpoint
	deviceDisplays map[string]*deviceListEntry
	Display        fyne.CanvasObject
	container      *fyne.Container
}

func newDeviceList(main *mainview.MainView, ep *rpc.RpcEndpoint) *deviceList {
	cont := container.NewVBox()
	return &deviceList{
		main:           main,
		ep:             ep,
		deviceDisplays: make(map[string]*deviceListEntry),
		Display:        container.NewVScroll(cont),
		container:      cont,
	}
}

func (d *deviceList) Set(key string, dev rpc.DeviceInfo) {
	disp, update := d.deviceDisplays[key]

	if update {
		log.Printf("updating display for %s", key)
	} else {
		log.Printf("adding display for %s", key)
		disp = newDeviceListEntry(d.main, d.ep, dev)
		d.deviceDisplays[key] = disp
		d.container.Add(disp.container)
		d.container.Refresh()
		disp.container.Refresh()
	}
	disp.Update(dev)
}

func (d *deviceList) Remove(key string) {
	delete(d.deviceDisplays, key)
	newList := make([]fyne.CanvasObject, 0, len(d.deviceDisplays))
	for _, disp := range d.deviceDisplays {
		newList = append(newList, disp.container)
	}
	d.container.Objects = newList
	d.container.Refresh()
}

type deviceListEntry struct {
	container *fyne.Container
	icon      *widget.Icon
	name      *widget.Label
	status    *widget.Label
}

func newDeviceListEntry(main *mainview.MainView, ep *rpc.RpcEndpoint, device rpc.DeviceInfo) *deviceListEntry {
	entry := &deviceListEntry{}

	entry.icon = widget.NewIcon(theme.ComputerIcon())
	entry.icon.Resize(fyne.Size{Width: 64, Height: 64})

	entry.name = widget.NewLabel("")

	entry.status = widget.NewLabel("")

	entry.container = container.NewHBox(entry.icon, entry.name, entry.status,
		layout.NewSpacer(),
		widget.NewButton("Select", func() {
			main.PushView(newDeviceView(ep, device))
		}),
	)

	return entry
}

func (d *deviceListEntry) Update(device rpc.DeviceInfo) {
	d.name.SetText(device.Name())

	var status string
	if device.Online {
		status = "Online"
	} else {
		status = "Offline"
	}

	d.status.SetText(status)
	d.container.Refresh()
}
