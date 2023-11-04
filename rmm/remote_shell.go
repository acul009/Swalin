package rmm

import (
	"fmt"
	"io"
	"log"
	"os"
	"rahnit-rmm/rpc"
	"syscall"

	"github.com/ActiveState/termtest/conpty"
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
	cpty, err := conpty.New(80, 25)
	if err != nil {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 500,
			Msg:  "Unable to create conpty",
		})
		return fmt.Errorf("error creating conpty: %w", err)
	}

	pid, _, err := cpty.Spawn(
		"C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0\\powershell.exe",
		[]string{},
		&syscall.ProcAttr{
			Env: os.Environ(),
		},
	)
	if err != nil {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 500,
			Msg:  "Unable to start shell",
		})
		return fmt.Errorf("error starting shell: %w", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("error finding process: %w", err)
	}

	go func() {
		_, err := process.Wait()
		if err != nil {
			log.Fatalf("Error waiting for process: %v", err)
		}
		cpty.Close()
	}()

	session.WriteResponseHeader(rpc.SessionResponseHeader{
		Code: 200,
		Msg:  "OK",
	})

	errChan := make(chan error)

	go func() {
		_, err = io.Copy(cpty.InPipe(), session)
		errChan <- err
	}()

	go func() {
		_, err = io.Copy(session, cpty.OutPipe())
		errChan <- err
	}()

	err = <-errChan
	if err != nil {
		return fmt.Errorf("error copying: %w", err)
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
