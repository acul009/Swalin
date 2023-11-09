//go:build windows
// +build windows

package rmm

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/winservices"
)

type windowsServiceSystem struct {
}

func GetServiceSystem() (ServiceSystem, error) {
	return &windowsServiceSystem{}, nil
}

func (s *windowsServiceSystem) GetStats() (*ServiceStats, error) {
	services, err := winservices.ListServices()
	if err != nil {
		return nil, fmt.Errorf("error listing services: %w", err)
	}

	infos := make([]ServiceInfo, 0, len(services))

	for _, service := range services {
		infos = append(infos, ServiceInfo{
			Name: service.Name,
		})
	}

	return &ServiceStats{Services: infos}, nil
}
