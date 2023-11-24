package rpc

import (
	"fmt"
	"log"
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
			payload := time.Now().UnixMicro()
			err := WriteMessage[int64](session, payload)
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
	log.Printf("Starting echo server")
	err := session.WriteResponseHeader(SessionResponseHeader{
		Code: 200,
		Msg:  "OK",
	})
	if err != nil {
		return fmt.Errorf("error writing response header: %w", err)
	}
	log.Printf("Sent Response")
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
