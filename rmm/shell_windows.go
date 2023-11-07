//go:build windows
// +build windows

package rmm

import (
	"fmt"
	"io"
	"log"
	"os"
	"syscall"

	"github.com/ActiveState/termtest/conpty"
)

type windowsShell struct {
	cpty    *conpty.ConPty
	outPipe io.ReadCloser
}

func startShell() (io.ReadWriteCloser, error) {
	cpty, err := conpty.New(80, 25)
	if err != nil {
		return nil, fmt.Errorf("error creating conpty: %w", err)
	}

	pid, _, err := cpty.Spawn(
		"C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0\\powershell.exe",
		[]string{},
		&syscall.ProcAttr{
			Env: os.Environ(),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error starting shell: %w", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("error finding process: %w", err)
	}

	go func() {
		_, err := process.Wait()
		if err != nil {
			log.Fatalf("Error waiting for process: %v", err)
		}
		cpty.Close()
	}()

	return &windowsShell{
		cpty:    cpty,
		outPipe: cpty.OutPipe(),
	}, nil
}

func (c *windowsShell) Read(p []byte) (n int, err error) {
	return c.outPipe.Read(p)
}

func (c *windowsShell) Write(p []byte) (n int, err error) {
	bytes, err := c.cpty.Write(p)
	return int(bytes), err
}

func (c *windowsShell) Close() error {
	var retErr error

	err := c.outPipe.Close()
	if err != nil {
		retErr = err
	}

	err = c.cpty.Close()
	if err != nil {
		retErr = err
	}

	return retErr
}
