package rmm

type ServiceSystem interface {
	GetStats() (*ServiceStats, error)
}

type ServiceStatus int

const (
	ServiceStatusStarting ServiceStatus = iota
	ServiceStatusRunning
	ServiceStatusStopping
	ServiceStatusStopped
	ServiceStatusError
	ServiceStatusUnknown
)

type ServiceInfo struct {
	Name        string
	Description string
	Enabled     bool
	Status      ServiceStatus
}

type ServiceStats struct {
	Services []ServiceInfo
}
