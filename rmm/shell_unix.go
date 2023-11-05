//go:build !windows
// +build !windows

package rmm

import (
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"
)

func startShell() (io.ReadWriteCloser, error) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "bash"
	}

	env := os.Environ()
	env = append(env, "TERM=xterm-256color")
	c := exec.Command(shell)
	c.Env = env

	// Start the command with a pty.
	f, err := pty.Start(c)
	return f, err
}
