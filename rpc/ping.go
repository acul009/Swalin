package rpc

import (
	"fmt"
	"time"
)

func PingHandler() RpcCommand {
	return &PingCmd{}
}

type PingCmd struct {
}

func (p *PingCmd) ExecuteClient(session *RpcSession) error {
	var errorOccured error = nil

	fmt.Println("Pinging...")

	go func() {
		for errorOccured == nil {
			time.Sleep(time.Second)
			payload := fmt.Sprintf("%d\n", time.Now().UnixMicro())
			_, err := session.Write([]byte(payload))
			if err != nil {
				errorOccured = err
				return
			}
		}
	}()

	for errorOccured == nil {
		var timestamp int64
		err := ReadMessage[*int64](session, &timestamp)
		if err != nil {
			errorOccured = err
			break
		}

		difference := time.Now().UnixMicro() - timestamp
		fmt.Printf("Time difference: %d\n", difference)
	}

	return errorOccured
}

func (p *PingCmd) ExecuteServer(session *RpcSession) error {
	session.WriteResponseHeader(SessionResponseHeader{
		Code: 200,
		Msg:  "OK",
	})
	for {
		var timestamp int64
		err := ReadMessage[*int64](session, &timestamp)
		if err != nil {
			return err
		}
		WriteMessage[*int64](session, &timestamp)
	}
}

func (p *PingCmd) GetKey() string {
	return "ping"
}
