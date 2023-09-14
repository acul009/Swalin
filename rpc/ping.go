package rpc

import (
	"fmt"
	"strconv"
	"time"
)

func PingHandler() *PingCmd {
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
		data, err := session.ReadUntil([]byte("\n"), 17, 65536)
		if err != nil {
			errorOccured = err
			return err
		}
		//parse data as number string
		timestamp, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			errorOccured = err
			return err
		}

		difference := time.Now().UnixMicro() - timestamp
		fmt.Printf("Time difference: %d\n", difference)
	}

	return errorOccured
}

func (p *PingCmd) ExecuteServer(session *RpcSession) error {
	for {
		data, err := session.ReadUntil([]byte("\n"), 17, 65536)
		if err != nil {
			return err
		}
		session.Write(append(data, []byte("\n")...))
	}
}

func (p *PingCmd) GetKey() string {
	return "ping"
}
