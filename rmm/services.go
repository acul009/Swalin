package rmm

type ServiceSystem interface {
	ListServices() ([]ServiceInfo, error)
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
