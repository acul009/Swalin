package rpc

// import (
// 	"github.com/rahn-it/svalin/util"
// )

// type getPendingEnrollmentsCommand struct {
// 	*SyncDownCommand[string, Enrollment]
// }

// func GetPendingEnrollmentsHandler() RpcCommand {
// 	return &getPendingEnrollmentsCommand{
// 		SyncDownCommand: NewSyncDownCommand[string, Enrollment](nil),
// 	}
// }

// func NewGetPendingEnrollmentsCommand(targetMap util.ObservableMap[string, Enrollment]) *getPendingEnrollmentsCommand {
// 	return &getPendingEnrollmentsCommand{
// 		SyncDownCommand: NewSyncDownCommand[string, Enrollment](targetMap),
// 	}
// }

// func (c *getPendingEnrollmentsCommand) GetKey() string {
// 	return "get-pending-enrollments"
// }

// func (c *getPendingEnrollmentsCommand) ExecuteServer(session *RpcSession) error {
// 	devicemap := util.NewObservableMap[string, Enrollment]()

// 	for pub, enrollment := range session.connection.server.enrollment.getAll() {
// 		devicemap.Set(pub, enrollment)
// 	}

// 	unsubscribe := session.connection.server.enrollment.subscribe(
// 		func(key string, value Enrollment) {
// 			devicemap.Set(key, value)
// 		},
// 		func(key string) {
// 			devicemap.Delete(key)
// 		},
// 	)
// 	defer unsubscribe()

// 	c.SyncDownCommand.targetMap = devicemap
// 	return c.SyncDownCommand.ExecuteServer(session)
// }
