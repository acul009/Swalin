package rmm

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type systemdServiceSystem struct {
}

func getSystemdServiceSystem() ServiceSystem {
	return &systemdServiceSystem{}
}

type systemdUnit struct {
	Name        string `json:"unit"`
	Load        string `json:"load"`
	Active      string `json:"active"`
	Sub         string `json:"sub"`
	Description string `json:"description"`
}

func (s *systemdServiceSystem) GetStats() (*ServiceStats, error) {
	cmd := exec.Command("systemctl", "--no-pager", "list-units", "--output=json", "--type=service", "--all")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list units: %w", err)
	}

	var units []systemdUnit = make([]systemdUnit, 0)
	err = json.Unmarshal(output, &units)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	services := make([]ServiceInfo, 0, len(units))

	for _, unit := range units {
		var status ServiceStatus

		switch unit.Sub {
		case "active", "running", "listening":
			status = ServiceStatusRunning
		case "dead", "exited":
			status = ServiceStatusStopped
		case "waiting":
			status = ServiceStatusRunning
		default:
			status = ServiceStatusUnknown
		}

		services = append(services, ServiceInfo{
			Name:        unit.Name,
			Description: unit.Description,
			Enabled:     unit.Active == "active",
			Status:      status,
		})
	}

	return &ServiceStats{Services: services}, nil
}
