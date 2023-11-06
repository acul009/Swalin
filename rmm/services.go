package rmm

type ServiceSystem interface {
	ListServices() ([]ServiceInfo, error)
}

type ServiceInfo struct {
	Name        string
	Pid         int32
	Description string
}
