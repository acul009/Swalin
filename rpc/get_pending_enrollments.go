package rpc

import (
	"fmt"
	"rahnit-rmm/util"
)

func NewGetPendingEnrollmentsCommand(targetMap *util.ObservableMap[string, Enrollment]) *getPendingEnrollmentsCommand {
	return &getPendingEnrollmentsCommand{
		enrollments: targetMap,
	}
}

type getPendingEnrollmentsCommand struct {
	enrollments *util.ObservableMap[string, Enrollment]
}

func GetPendingEnrollmentsHandler() RpcCommand {
	return &getPendingEnrollmentsCommand{
		enrollments: nil,
	}
}

type enrollmentUpdate struct {
	Key        string
	Enrollment *Enrollment
}

func (c *getPendingEnrollmentsCommand) GetKey() string {
	return "get-pending-enrollments"
}

func (c *getPendingEnrollmentsCommand) ExecuteClient(session *RpcSession) error {
	list := make(map[string]Enrollment)

	err := ReadMessage[map[string]Enrollment](session, list)
	if err != nil {
		return fmt.Errorf("error reading message: %w", err)
	}

	for key, enrollment := range list {
		c.enrollments.Set(key, enrollment)
	}

	for {
		update := &enrollmentUpdate{}
		err := ReadMessage[*enrollmentUpdate](session, update)
		if err != nil {
			return fmt.Errorf("error reading message: %w", err)
		}

		if update.Enrollment != nil {
			c.enrollments.Set(update.Key, *update.Enrollment)
		} else {
			c.enrollments.Delete(update.Key)
		}
	}

}

func (c *getPendingEnrollmentsCommand) ExecuteServer(session *RpcSession) error {
	session.WriteResponseHeader(SessionResponseHeader{
		Code: 200,
		Msg:  "OK",
	})

	err := WriteMessage[map[string]Enrollment](session, session.connection.server.enrollment.getAll())
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}

	updateErrChan := make(chan error)

	unsubscribe := session.connection.server.enrollment.subscribe(
		func(key string, enrollment Enrollment) {
			update := &enrollmentUpdate{
				Key:        key,
				Enrollment: &enrollment,
			}

			err := WriteMessage[*enrollmentUpdate](session, update)
			if err != nil {
				updateErrChan <- fmt.Errorf("error writing update message: %w", err)
			}
		},
		func(key string) {
			update := &enrollmentUpdate{
				Key:        key,
				Enrollment: nil,
			}

			err := WriteMessage[*enrollmentUpdate](session, update)
			if err != nil {
				updateErrChan <- fmt.Errorf("error writing remove message: %w", err)
			}
		},
	)

	err = <-updateErrChan
	unsubscribe()
	return fmt.Errorf("error subscribing to updates: %w", err)
}
