package system

import "github.com/rahn-it/svalin/pki"

type DeviceInfo struct {
	Certificate *pki.Certificate
	LiveInfo    LiveDeviceInfo
}

type LiveDeviceInfo struct {
	Online bool
}
