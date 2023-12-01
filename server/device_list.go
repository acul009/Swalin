package server

import (
	"fmt"

	"github.com/rahn-it/svalin/db"
	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/util"
	"go.etcd.io/bbolt"
)

type DeviceInfo struct {
	Certificate *pki.Certificate
	liveInfo    LiveDeviceInfo
}

type LiveDeviceInfo struct {
	Online bool
}

var _ util.ObservableMap[string, *DeviceInfo] = (*DeviceList)(nil)

type DeviceList struct {
	observerHandler *util.MapObserverHandler[string, *DeviceInfo]
	scope           db.Scope
	online          map[string]bool
}

func NewDeviceList(scope db.Scope) *DeviceList {
	return &DeviceList{
		observerHandler: util.NewMapObserverHandler[string, *DeviceInfo](),
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

func (d *DeviceList) Get(key string) (*DeviceInfo, bool) {
	var raw []byte
	err := d.scope.View(func(b *bbolt.Bucket) error {
		raw = b.Get([]byte(key))
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

	return &DeviceInfo{
		Certificate: cert,
		liveInfo: LiveDeviceInfo{
			Online: d.isOnline(key),
		},
	}, true
}

func (d *DeviceList) ForEach(handler func(key string, value *DeviceInfo) error) error {
	return d.scope.View(func(b *bbolt.Bucket) error {
		return b.ForEach(func(k, v []byte) error {
			cert, err := pki.CertificateFromPem(v)
			if err != nil {
				return err
			}

			di := &DeviceInfo{
				Certificate: cert,
				liveInfo: LiveDeviceInfo{
					Online: d.isOnline(cert.GetPublicKey().Base64Encode()),
				},
			}

			return handler(cert.GetPublicKey().Base64Encode(), di)
		})
	})
}

func (d *DeviceList) Subscribe(onUpdate func(string, *DeviceInfo), onRemove func(string, *DeviceInfo)) func() {
	return d.observerHandler.Subscribe(onUpdate, onRemove)
}

func (d *DeviceList) AddDeviceToDB(cert *pki.Certificate) error {
	// TODO: verify certificate

	key := cert.GetPublicKey().Base64Encode()
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

	di := &DeviceInfo{
		Certificate: cert,
		liveInfo: LiveDeviceInfo{
			Online: d.isOnline(key),
		},
	}

	d.observerHandler.NotifyUpdate(key, di)

	return nil
}
