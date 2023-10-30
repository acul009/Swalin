package mainview

import (
	"context"
	"rahnit-rmm/rpc"
	"rahnit-rmm/util"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type enrollmentView struct {
	ep          *rpc.RpcEndpoint
	enrollments *util.ObservableMap[string, rpc.Enrollment]
	needsUpdate bool
	mutex       sync.Mutex
	visible     bool
}

func newEnrollmentView(ep *rpc.RpcEndpoint) *enrollmentView {

	e := &enrollmentView{
		ep:          ep,
		enrollments: util.NewObservableMap[string, rpc.Enrollment](),
		needsUpdate: false,
		mutex:       sync.Mutex{},
		visible:     false,
	}

	updateCommand := rpc.NewGetPendingEnrollmentsCommand(e.enrollments)

	go func() {
		err := e.ep.SendCommand(context.Background(), updateCommand)
		if err != nil {
			panic(err)
		}
	}()

	return e
}

func (e *enrollmentView) Prepare() fyne.CanvasObject {
	e.mutex.Lock()
	e.visible = true
	e.needsUpdate = true
	e.mutex.Unlock()

	values := e.enrollments.Values()

	e.enrollments.Subscribe(
		func(key string, enrollment rpc.Enrollment) {
			e.mutex.Lock()
			defer e.mutex.Unlock()
			e.needsUpdate = true
		},
		func(key string) {
			e.mutex.Lock()
			defer e.mutex.Unlock()
			e.needsUpdate = true
		},
	)

	list := widget.NewList(
		func() int {
			return len(values)
		},
		func() fyne.CanvasObject {
			return container.NewVBox(
				container.NewHBox(
					widget.NewLabel("Address"),
					widget.NewLabel("Request Time"),
				),
				widget.NewLabel("PubKey\n\n\n\n\n"),
				widget.NewButton("Enroll", nil),
			)
		},
		func(i int, o fyne.CanvasObject) {
			enrollment := values[i]

			pubKey := enrollment.PublicKey.PemEncode()

			grid := o.(*fyne.Container)
			grid.Objects[0].(*fyne.Container).Objects[0].(*widget.Label).SetText(enrollment.Addr)
			grid.Objects[0].(*fyne.Container).Objects[1].(*widget.Label).SetText(enrollment.RequestTime.Format("2006-01-02 15:04:05"))
			grid.Objects[1].(*widget.Label).SetText(string(pubKey))
			// pubKeyDisp := grid.Objects[2].(*widget.Entry)
			// pubKeyDisp.OnChanged = func(s string) {
			// 	pubKeyDisp.Text =
			// }
		},
	)

	go func() {
		for e.visible {
			time.Sleep(time.Second)
			e.mutex.Lock()
			if e.needsUpdate {
				values = e.enrollments.Values()
				e.needsUpdate = false
				e.mutex.Unlock()
				list.Refresh()
			} else {
				e.mutex.Unlock()
			}
		}
	}()

	return list
}

func (e *enrollmentView) Close() {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.visible = false
}
