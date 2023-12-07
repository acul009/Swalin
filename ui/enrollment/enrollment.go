package enrollment

// import (
// 	"context"
// 	"log"
// 	"github.com/rahn-it/svalin/pki"
// 	"github.com/rahn-it/svalin/rmm"
// 	"github.com/rahn-it/svalin/rpc"
// 	"github.com/rahn-it/svalin/ui/mainview.go"
// 	"github.com/rahn-it/svalin/util"
// 	"sync"
// 	"time"

// 	"fyne.io/fyne/v2"
// 	"fyne.io/fyne/v2/container"
// 	"fyne.io/fyne/v2/theme"
// 	"fyne.io/fyne/v2/widget"
// )

// type enrollmentView struct {
// 	widget.BaseWidget
// 	main        *mainview.MainView
// 	cli         *client.Client
// 	credentials *pki.PermanentCredentials
// 	enrollments util.ObservableMap[string, rpc.Enrollment]
// 	needsUpdate bool
// 	mutex       sync.Mutex
// 	visible     bool
// }

// func NewEnrollmentView(main *mainview.MainView, cli *client.Client, credentials *pki.PermanentCredentials) *enrollmentView {

// 	e := &enrollmentView{
// 		main:        main,
// 		cli:         cli,
// 		credentials: credentials,
// 		enrollments: util.NewObservableMap[string, rpc.Enrollment](),
// 		needsUpdate: false,
// 		mutex:       sync.Mutex{},
// 		visible:     false,
// 	}

// 	e.ExtendBaseWidget(e)

// 	updateCommand := rpc.NewGetPendingEnrollmentsCommand(e.enrollments)

// 	running, err := cli.SendCommand(context.Background(), updateCommand)
// 	if err != nil {
// 		panic(err)
// 	}

// 	go func() {
// 		err := running.Wait()
// 		if err != nil {
// 			panic(err)
// 		}
// 	}()

// 	return e
// }

// func (e *enrollmentView) Icon() fyne.Resource {
// 	return theme.FolderNewIcon()
// }

// func (e *enrollmentView) Name() string {
// 	return "Enrollments"
// }

// func (e *enrollmentView) Prepare() fyne.CanvasObject {
// 	e.mutex.Lock()
// 	e.visible = true
// 	e.needsUpdate = true
// 	e.mutex.Unlock()

// 	all := e.enrollments.GetAll()
// 	values := make([]rpc.Enrollment, 0, len(all))
// 	for _, v := range all {
// 		values = append(values, v)
// 	}

// 	e.enrollments.Subscribe(
// 		func(key string, enrollment rpc.Enrollment) {
// 			e.mutex.Lock()
// 			defer e.mutex.Unlock()
// 			e.needsUpdate = true
// 		},
// 		func(key string) {
// 			e.mutex.Lock()
// 			defer e.mutex.Unlock()
// 			e.needsUpdate = true
// 		},
// 	)

// 	list := widget.NewList(
// 		func() int {
// 			return len(values)
// 		},
// 		func() fyne.CanvasObject {
// 			form := widget.NewForm(
// 				widget.NewFormItem("Name", widget.NewEntry()),
// 			)
// 			form.SubmitText = "Enroll"
// 			form.OnSubmit = func() {}

// 			return container.NewVBox(
// 				container.NewHBox(
// 					widget.NewLabel("Address"),
// 					widget.NewLabel("Request Time"),
// 				),
// 				widget.NewLabel("PubKey\n\n\n\n\n"),
// 				form,
// 			)
// 		},
// 		func(i int, o fyne.CanvasObject) {
// 			enrollment := values[i]

// 			pubKey := enrollment.PublicKey.PemEncode()

// 			grid := o.(*fyne.Container)
// 			grid.Objects[0].(*fyne.Container).Objects[0].(*widget.Label).SetText(enrollment.Addr)
// 			grid.Objects[0].(*fyne.Container).Objects[1].(*widget.Label).SetText(enrollment.RequestTime.Format("2006-01-02 15:04:05"))
// 			grid.Objects[1].(*widget.Label).SetText(string(pubKey))
// 			form := grid.Objects[2].(*widget.Form)
// 			nameEntry := form.Items[0].Widget.(*widget.Entry)

// 			form.OnSubmit = func() {
// 				name := nameEntry.Text
// 				log.Printf("Enrolling %s", name)

// 				if name == "" {
// 					return
// 				}

// 				cert, err := pki.CreateAgentCert(name, enrollment.PublicKey, e.credentials)
// 				if err != nil {
// 					panic(err)
// 				}

// 				err = e.cli.SendSyncCommand(context.Background(), rpc.NewEnrollAgentCommand(cert))
// 				if err != nil {
// 					panic(err)
// 				}

// 			}
// 			form.Refresh()

// 			// pubKeyDisp := grid.Objects[2].(*widget.Entry)
// 			// pubKeyDisp.OnChanged = func(s string) {
// 			// 	pubKeyDisp.Text =
// 			// }
// 		},
// 	)

// 	go func() {
// 		for e.visible {
// 			time.Sleep(time.Second)
// 			e.mutex.Lock()
// 			if e.needsUpdate {
// 				all := e.enrollments.GetAll()
// 				values = make([]rpc.Enrollment, 0, len(all))
// 				for _, v := range all {
// 					values = append(values, v)
// 				}
// 				e.needsUpdate = false
// 				e.mutex.Unlock()
// 				list.Refresh()
// 			} else {
// 				e.mutex.Unlock()
// 			}
// 		}
// 	}()

// 	return list
// }

// func (e *enrollmentView) Close() {
// 	e.mutex.Lock()
// 	defer e.mutex.Unlock()
// 	e.visible = false
// }

// func (e *enrollmentView) CreateRenderer() fyne.WidgetRenderer {
// 	return &enrollmentViewRenderer{
// 		widget: e,
// 	}
// }

// type enrollmentViewRenderer struct {
// 	widget *enrollmentView
// }

// func (e *enrollmentViewRenderer) Layout(size fyne.Size) {

// }

// func (e *enrollmentViewRenderer) MinSize() fyne.Size {
// 	return fyne.NewSize(0, 0)
// }

// func (e *enrollmentViewRenderer) Refresh() {

// }

// func (e *enrollmentViewRenderer) Destroy() {

// }

// func (e *enrollmentViewRenderer) Objects() []fyne.CanvasObject {
// 	return nil
// }
