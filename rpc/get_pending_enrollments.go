package rpc

import (
	"log"
	"rahnit-rmm/util"
)

type getPendingEnrollmentsCommand struct {
	*syncDownCommand[string, Enrollment]
}

func GetPendingEnrollmentsHandler() RpcCommand {
	return &getPendingEnrollmentsCommand{
		syncDownCommand: NewSyncDownCommand[string, Enrollment](nil),
	}
}

func NewGetPendingEnrollmentsCommand(targetMap *util.ObservableMap[string, Enrollment]) *getPendingEnrollmentsCommand {
	return &getPendingEnrollmentsCommand{
		syncDownCommand: NewSyncDownCommand[string, Enrollment](targetMap),
	}
}

func (c *getPendingEnrollmentsCommand) GetKey() string {
	return "get-pending-enrollments"
}

func (c *getPendingEnrollmentsCommand) ExecuteServer(session *RpcSession) error {
	devicemap := util.NewObservableMap[string, Enrollment]()

	for pub, enrollment := range session.connection.server.enrollment.getAll() {
		log.Printf("enrollment initial added: %s", pub)
		devicemap.Set(pub, enrollment)
	}

	unsubscribe := session.connection.server.enrollment.subscribe(
		func(key string, value Enrollment) {
			log.Printf("enrollment added: %s", key)
			devicemap.Set(key, value)
		},
		func(key string) {
			log.Printf("enrollment removed: %s", key)
			devicemap.Delete(key)
		},
	)
	defer unsubscribe()

	c.syncDownCommand.targetMap = devicemap
	return c.syncDownCommand.ExecuteServer(session)
}
