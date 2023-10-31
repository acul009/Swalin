package managment

import (
	"rahnit-rmm/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type deviceList struct {
	deviceDisplays map[string]*fyne.Container
	Display        fyne.CanvasObject
	container      *fyne.Container
}

func newDeviceList() *deviceList {
	cont := container.NewVBox()
	return &deviceList{
		deviceDisplays: make(map[string]*fyne.Container),
		Display:        container.NewVScroll(cont),
		container:      cont,
	}
}

func (d *deviceList) Set(key string, dev rpc.DeviceInfo) {
	currentDisplay, update := d.deviceDisplays[key]

	icon := widget.NewIcon(theme.ComputerIcon())
	icon.Resize(fyne.Size{Width: 64, Height: 64})

	var status string
	if dev.Online {
		status = "Online"
	} else {
		status = "Offline"
	}

	disp := container.NewHBox(
		icon,
		widget.NewLabel(status),
		widget.NewLabel(dev.Name()),
	)

	if update {
		currentDisplay.Objects = []fyne.CanvasObject{disp}
		currentDisplay.Refresh()
	} else {
		newList := append(d.container.Objects, disp)
		d.container.Objects = newList
		d.container.Refresh()
	}
}

func (d *deviceList) Remove(key string) {
	delete(d.deviceDisplays, key)
	newList := make([]fyne.CanvasObject, 0, len(d.deviceDisplays))
	for _, disp := range d.deviceDisplays {
		newList = append(newList, disp.Objects...)
	}
	d.container.Objects = newList
	d.container.Refresh()
}
