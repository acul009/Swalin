package rpc

import (
	"context"
	"fmt"
	"rahnit-rmm/config"
	"rahnit-rmm/pki"
	"rahnit-rmm/util"
)

type DeviceInfo struct {
	Certificate *pki.Certificate
	online      bool
}

type DeviceList struct {
	devices *util.ObservableMap[string, DeviceInfo]
}

func NewDeviceListFromDB() (*DeviceList, error) {
	d := &DeviceList{
		devices: util.NewObservableMap[string, DeviceInfo](),
	}

	db := config.DB()
	devices, err := db.Device.Query().All(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}

	for _, device := range devices {
		cert, err := pki.CertificateFromPem([]byte(device.Certificate))
		if err != nil {
			return nil, fmt.Errorf("failed to parse device certificate: %w", err)
		}
		d.devices.Set(cert.GetPublicKey().Base64Encode(), DeviceInfo{
			Certificate: cert,
			online:      false,
		})
	}

	return d, nil
}

func (d *DeviceList) AddDeviceToDB(cert *pki.Certificate) error {
	if d.devices.Has(cert.GetPublicKey().Base64Encode()) {
		return fmt.Errorf("device already exists")
	}

	db := config.DB()
	err := db.Device.Create().SetPublicKey(cert.GetPublicKey().Base64Encode()).SetCertificate(string(cert.PemEncode())).Exec(context.Background())
	if err != nil {
		return fmt.Errorf("failed to create device: %w", err)
	}

	d.devices.Set(cert.GetPublicKey().Base64Encode(), DeviceInfo{
		Certificate: cert,
		online:      false,
	})

	return nil
}

func (d *DeviceList) Subscribe(onSet func(string, DeviceInfo), onRemove func(string)) func() {
	return d.devices.Subscribe(onSet, onRemove)
}

func (d *DeviceList) UpdateDeviceStatus(pubKey string, update func(device DeviceInfo) DeviceInfo) {
	d.devices.Update(pubKey, update)
}

func (d *DeviceList) GetAll() map[string]DeviceInfo {
	return d.devices.GetAll()
}
