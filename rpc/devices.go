package rpc

import (
	"context"
	"fmt"
	"log"
	"rahnit-rmm/config"
	"rahnit-rmm/pki"
	"rahnit-rmm/util"
)

type DeviceInfo struct {
	Certificate *pki.Certificate
	Online      bool
}

func (d DeviceInfo) Name() string {
	return d.Certificate.GetName()
}

type DeviceList struct {
	devices util.ObservableMap[string, *DeviceInfo]
}

func NewDeviceListFromDB() (*DeviceList, error) {
	d := &DeviceList{
		devices: util.NewObservableMap[string, *DeviceInfo](),
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
		d.devices.Set(cert.GetPublicKey().Base64Encode(), &DeviceInfo{
			Certificate: cert,
			Online:      false,
		})
	}

	return d, nil
}

func (d *DeviceList) AddDeviceToDB(cert *pki.Certificate) error {
	_, ok := d.devices.Get(cert.GetPublicKey().Base64Encode())
	if ok {
		return fmt.Errorf("device already exists")
	}

	db := config.DB()
	err := db.Device.Create().SetPublicKey(cert.GetPublicKey().Base64Encode()).SetCertificate(string(cert.PemEncode())).Exec(context.Background())
	if err != nil {
		return fmt.Errorf("failed to create device: %w", err)
	}

	d.devices.Set(cert.GetPublicKey().Base64Encode(), &DeviceInfo{
		Certificate: cert,
		Online:      false,
	})

	return nil
}

func (d *DeviceList) Subscribe(onSet func(string, *DeviceInfo), onRemove func(string)) func() {
	return d.devices.Subscribe(onSet, onRemove)
}

func (d *DeviceList) UpdateDeviceStatus(pubKey string, update func(device *DeviceInfo) *DeviceInfo) {
	log.Printf("Updating device status for %s", pubKey)
	d.devices.Update(pubKey,
		func(device *DeviceInfo, found bool) (*DeviceInfo, bool) {
			if !found {
				log.Printf("Unknown device login: %s", pubKey)
				return nil, false
			}

			device = update(device)
			return device, true
		},
	)
}

func (d *DeviceList) GetAll() map[string]*DeviceInfo {
	return d.devices.GetAll()
}
