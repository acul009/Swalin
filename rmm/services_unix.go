//go:build !windows
// +build !windows

package rmm

import (
	"fmt"
	"os/exec"
)

var serviceSystem ServiceSystem

func GetServiceSystem() (ServiceSystem, error) {
	if serviceSystem != nil {
		return serviceSystem, nil
	}

	cmd := exec.Command("systemctl", "--version")
	err := cmd.Start()
	if err != nil {
		serviceSystem = getSystemdServiceSystem()
		return serviceSystem, nil
	}

	return nil, fmt.Errorf("no compatible service system found")
}
