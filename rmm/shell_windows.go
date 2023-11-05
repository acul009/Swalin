//go:build windows
// +build windows

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

type windowsShell struct {
}

func startShell() io.ReadWriteCloser {
	cpty, err := conpty.New(80, 25)
	if err != nil {
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
}
