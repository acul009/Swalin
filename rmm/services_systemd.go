package rmm

import (
	"os/exec"
	"strings"
)

type systemdServiceSystem struct {
}

func getSystemdServiceSystem() ServiceSystem {
	return &systemdServiceSystem{}
}

func (s *systemdServiceSystem) ListServices() ([]ServiceInfo, error) {
	cmd := exec.Command("systemctl", "--no-pager", "list-units", "--type=service", "--all")

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	services := make([]ServiceInfo, 0, len(lines))

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		service := ServiceInfo{
			Name:        fields[0],
			Description: fields[4],
		}

		services = append(services, service)
	}

	return services, nil
}
