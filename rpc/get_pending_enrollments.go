package rpc

type getPendingEnrollmentsCommand struct {
}

func (c *getPendingEnrollmentsCommand) ExecuteClient(session *RpcSession) error {
	return nil
}

func (c *getPendingEnrollmentsCommand) ExecuteServer(session *RpcSession) error {
	session.WriteResponseHeader(SessionResponseHeader{
		Code: 200,
		Msg:  "OK",
	})

	activeEnrollments := session.connection.server.enrollment.list()

	WriteMessage[[]Enrollment](session, activeEnrollments)
	return nil
}
