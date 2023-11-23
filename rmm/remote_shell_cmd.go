package rmm

import (
	"fmt"
	"io"
	"rahnit-rmm/rpc"
)

func RemoteShellCommandHandler() rpc.RpcCommand[*Dependencies] {
	return &remoteShellCommand{}
}

type remoteShellCommand struct {
	input  io.ReadCloser
	output io.WriteCloser
}

func NewRemoteShellCommand(input io.ReadCloser, output io.WriteCloser) *remoteShellCommand {
	return &remoteShellCommand{
		input:  input,
		output: output,
	}
}

func (cmd *remoteShellCommand) GetKey() string {
	return "remote-shell"
}

func (cmd *remoteShellCommand) ExecuteServer(session *rpc.RpcSession[*Dependencies]) error {
	shell, err := startShell()
	if err != nil {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 500,
			Msg:  "Unable to start shell",
		})
		return fmt.Errorf("error starting shell: %w", err)
	}

	session.WriteResponseHeader(rpc.SessionResponseHeader{
		Code: 200,
		Msg:  "OK",
	})

	errChan := make(chan error)

	go func() {
		_, err = io.Copy(shell, session)
		errChan <- err
	}()

	go func() {
		_, err = io.Copy(session, shell)
		errChan <- err
	}()

	err = <-errChan
	if err != nil {
		return fmt.Errorf("error copying: %w", err)
	}

	return nil
}

func (cmd *remoteShellCommand) ExecuteClient(session *rpc.RpcSession[*Dependencies]) error {
	errChan := make(chan error)
	go func() {
		_, err := io.Copy(session, cmd.input)
		errChan <- err
	}()
	go func() {
		_, err := io.Copy(cmd.output, session)
		errChan <- err
	}()

	return <-errChan
}
