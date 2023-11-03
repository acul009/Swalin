package rmm

import (
	"fmt"
	"io"
	"os/exec"
	"rahnit-rmm/rpc"
)

func RemoteShellCommandHandler() rpc.RpcCommand {
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

func (cmd *remoteShellCommand) ExecuteServer(session *rpc.RpcSession) error {
	shell := getShellCommand()

	shellCmd := exec.Command(shell[0], shell[1:]...)
	shellCmd.Stdin = session
	shellCmd.Stdout = session
	shellCmd.Stderr = session

	session.WriteResponseHeader(rpc.SessionResponseHeader{
		Code: 200,
		Msg:  "OK",
	})

	err := shellCmd.Start()
	if err != nil {
		return fmt.Errorf("error starting shell command: %w", err)
	}

	err = shellCmd.Wait()
	if err != nil {
		return fmt.Errorf("error waiting for shell command: %w", err)
	}

	return nil
}

func (cmd *remoteShellCommand) ExecuteClient(session *rpc.RpcSession) error {
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
