package server

import (
	"fmt"

	"github.com/rahn-it/svalin/db"
	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/system"
	"github.com/rahn-it/svalin/util"
	"go.etcd.io/bbolt"
)

var _ util.ObservableMap[string, *system.DeviceInfo] = (*DeviceList)(nil)

type DeviceList struct {
	observerHandler *util.MapObserverHandler[string, *system.DeviceInfo]
	scope           db.Scope
	online          map[string]bool
}

func NewDeviceList(scope db.Scope) *DeviceList {
	return &DeviceList{
		observerHandler: util.NewMapObserverHandler[string, *system.DeviceInfo](),
		scope:           scope,
	}
}

func (d *DeviceList) isOnline(key string) bool {
	online, ok := d.online[key]
	if !ok {
		return false
	}

	if ok && !online {
		delete(d.online, key)
	}
	return online
}

func (d *DeviceList) Get(key string) (*system.DeviceInfo, bool) {
	var raw []byte
	err := d.scope.View(func(b *bbolt.Bucket) error {
		val := b.Get([]byte(key))
		if val == nil {
			return nil
		}

		raw = make([]byte, len(val))
		copy(raw, val)
		return nil
	})
	if err != nil {
		panic(err)
	}

	if raw == nil {
		delete(d.online, key)
		return nil, false
	}

	cert, err := pki.CertificateFromPem(raw)
	if err != nil {
		panic(err)
	}

	return &system.DeviceInfo{
		Certificate: cert,
		LiveInfo: system.LiveDeviceInfo{
			Online: d.isOnline(key),
		},
	}, true
}

func (d *DeviceList) ForEach(handler func(key string, value *system.DeviceInfo) error) error {
	return d.scope.View(func(b *bbolt.Bucket) error {
		return b.ForEach(func(k, v []byte) error {
			cert, err := pki.CertificateFromPem(v)
			if err != nil {
				return err
			}

			di := &system.DeviceInfo{
				Certificate: cert,
				LiveInfo: system.LiveDeviceInfo{
					Online: d.isOnline(cert.PublicKey().Base64Encode()),
				},
			}

			return handler(cert.PublicKey().Base64Encode(), di)
		})
	})
}

func (d *DeviceList) Subscribe(onUpdate func(string, *system.DeviceInfo), onRemove func(string, *system.DeviceInfo)) func() {
	return d.observerHandler.Subscribe(onUpdate, onRemove)
}

func (d *DeviceList) AddDeviceToDB(cert *pki.Certificate) error {
	// TODO: verify certificate

	key := cert.PublicKey().Base64Encode()
	rawKey := []byte(key)
	pem := cert.PemEncode()

	err := d.scope.Update(func(b *bbolt.Bucket) error {
		raw := b.Get(rawKey)
		if raw != nil {
			return fmt.Errorf("device already exists")
		}
		return b.Put(rawKey, pem)
	})

	if err != nil {
		return fmt.Errorf("failed to add device: %w", err)
	}

	di := &system.DeviceInfo{
		Certificate: cert,
		LiveInfo: system.LiveDeviceInfo{
			Online: d.isOnline(key),
		},
	}

	d.observerHandler.NotifyUpdate(key, di)

	return nil
}

func (d *DeviceList) setOnlineStatus(key string, online bool) {
	if online {
		d.online[key] = true
	} else {
		delete(d.online, key)
	}
}
