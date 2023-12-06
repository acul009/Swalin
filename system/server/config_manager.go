package server

import (
	"fmt"

	"github.com/rahn-it/svalin/db"
	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/rmm"
	"github.com/rahn-it/svalin/util"
)

type ConfigManager struct {
	tunnelConfigHandler *HostConfigHandler[*rmm.TunnelConfig]
}

func NewConfigManager(verifier pki.Verifier, scope db.Scope) *ConfigManager {

	hostConfigScope := scope.Scope("host_configs")

	return &ConfigManager{
		tunnelConfigHandler: NewHostConfigHandler[*rmm.TunnelConfig](verifier, hostConfigScope),
	}
}

var _ util.ObservableMap[string, *pki.SignedArtifact[rmm.HostConfig]] = (*HostConfigHandler[rmm.HostConfig])(nil)

type HostConfigHandler[T rmm.HostConfig] struct {
	verifier        pki.Verifier
	db              db.Scope
	observerHandler *util.MapObserverHandler[string, *pki.SignedArtifact[T]]
}

func NewHostConfigHandler[T rmm.HostConfig](verifier pki.Verifier, scope db.Scope) *HostConfigHandler[T] {
	var t T

	db := scope.Scope(t.GetConfigKey())

	return &HostConfigHandler[T]{
		verifier:        verifier,
		db:              db,
		observerHandler: util.NewMapObserverHandler[string, *pki.SignedArtifact[T]](),
	}
}

func (h *HostConfigHandler[T]) ObserverCount() util.Observable[int] {
	return h.observerHandler.ObserverCount()
}

func (h *HostConfigHandler[T]) Subscribe(onSet func(string, *pki.SignedArtifact[T]), onRemove func(string, *pki.SignedArtifact[T])) func() {
	return h.observerHandler.Subscribe(onSet, onRemove)
}

func (h *HostConfigHandler[T]) Get(key string) (*pki.SignedArtifact[T], bool) {
	var raw []byte
	err := h.db.View(func(b db.Bucket) error {
		raw = b.Get([]byte(key))
		return nil
	})
	if err != nil {
		panic(err)
	}

	if raw == nil {
		return nil, false
	}

	artifact, err := pki.LoadSignedArtifact[T](raw, h.verifier)
	if err != nil {
		panic(err)
	}

	return artifact, true
}

func (h *HostConfigHandler[T]) ForEach(handler func(key string, value *pki.SignedArtifact[T]) error) error {
	return h.db.View(func(b db.Bucket) error {
		return b.ForEach(func(k, v []byte) error {
			artifact, err := pki.LoadSignedArtifact[T](v, h.verifier)
			if err != nil {
				return err
			}

			return handler(string(k), artifact)
		})
	})
}

func (h *HostConfigHandler[T]) UpdateConfig(config *pki.SignedArtifact[T]) error {
	pubKey := config.Artifact().GetHost().Base64Encode()
	pubKeyRaw := []byte(pubKey)

	err := h.db.Update(func(b db.Bucket) error {
		raw := b.Get(pubKeyRaw)

		if raw != nil {
			artifact, err := pki.LoadSignedArtifact[T](raw, h.verifier)
			if err != nil {
				return fmt.Errorf("error unmarshaling old config: %w", err)
			}

			if artifact.Timestamp() >= config.Timestamp() {
				return fmt.Errorf("old config received")
			}
		}

		err := b.Put(pubKeyRaw, config.Raw())
		if err != nil {
			return fmt.Errorf("error writing host config: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error updating host config: %w", err)
	}

	h.observerHandler.NotifyUpdate(pubKey, config)
	return nil
}
