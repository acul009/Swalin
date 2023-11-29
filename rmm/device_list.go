package rmm

import (
	"context"
	"fmt"
	"rahnit-rmm/config"
	"rahnit-rmm/pki"
	"rahnit-rmm/util"
)

type DeviceInfo struct {
	Certificate *pki.Certificate
	Online      bool
}

type DeviceList struct {
	util.UpdateableMap[string, *DeviceInfo]
}

func NewDeviceListFromDB() (*DeviceList, error) {
	d := &DeviceList{
		UpdateableMap: util.NewObservableMap[string, *DeviceInfo](),
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
		d.UpdateableMap.Set(cert.GetPublicKey().Base64Encode(), &DeviceInfo{
			Certificate: cert,
			Online:      false,
		})
	}

	return d, nil
}

func (d *DeviceList) AddDeviceToDB(cert *pki.Certificate) error {
	_, ok := d.UpdateableMap.Get(cert.GetPublicKey().Base64Encode())
	if ok {
		return fmt.Errorf("device already exists")
	}

	db := config.DB()
	err := db.Device.Create().SetPublicKey(cert.GetPublicKey().Base64Encode()).SetCertificate(string(cert.PemEncode())).Exec(context.Background())
	if err != nil {
		return fmt.Errorf("failed to create device: %w", err)
	}

	d.UpdateableMap.Set(cert.GetPublicKey().Base64Encode(), &DeviceInfo{
		Certificate: cert,
		Online:      false,
	})

	return nil
}
