//go:build !windows
// +build !windows

package rmm

import (
	"fmt"
	"io"
	"os/exec"
)

type unixShell struct {
	io.Writer
	io.Reader
	cmd *exec.Cmd
}

func startShell() (io.ReadWriteCloser, error) {
	cmd := exec.Command("/bin/bash")
	readInput, writeInput := io.Pipe()
	readOutput, writeOutput := io.Pipe()
	cmd.Stdin = readInput
	cmd.Stdout = writeOutput
	cmd.Stderr = writeOutput
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error starting shell: %w", err)
	}

	shell := &unixShell{
		Writer: writeInput,
		Reader: readOutput,
		cmd:    cmd,
	}

	go func() {
		err := cmd.Wait()
		if err != nil {
			panic(err)
		}
	}()

	return shell, nil
}

func (s *unixShell) Close() error {
	err := s.cmd.Process.Kill()
	defer s.cmd.Process.Release()
	if err != nil {
		return fmt.Errorf("error killing shell: %w", err)
	}

	return nil
}
