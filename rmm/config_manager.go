package rmm

import (
	"context"
	"fmt"
	"rahnit-rmm/ent"
	"rahnit-rmm/ent/device"
	"rahnit-rmm/ent/hostconfig"
	"rahnit-rmm/pki"
	"rahnit-rmm/util"
)

type ConfigManager struct {
	tunnelConfigHandler *HostConfigHandler[*TunnelConfig]
}

func NewConfigManager(verifier pki.Verifier, db *ent.Client) *ConfigManager {

	return &ConfigManager{
		tunnelConfigHandler: NewHostConfigHandler[*TunnelConfig](verifier, db),
	}
}

var _ util.ObservableMap[string, *pki.SignedArtifact[HostConfig]] = (*HostConfigHandler[HostConfig])(nil)

type HostConfigHandler[T HostConfig] struct {
	verifier        pki.Verifier
	db              *ent.Client
	configKey       string
	observerHandler *util.MapObserverHandler[string, *pki.SignedArtifact[T]]
}

func NewHostConfigHandler[T HostConfig](verifier pki.Verifier, db *ent.Client) *HostConfigHandler[T] {
	var t T

	return &HostConfigHandler[T]{
		verifier:        verifier,
		db:              db,
		configKey:       t.GetConfigKey(),
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
	dbEntry, err := h.db.HostConfig.Query().Where(
		hostconfig.HasDeviceWith(device.PublicKey(key)),
		hostconfig.Type(h.configKey),
	).Only(context.Background())

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, false
		}

		panic(err)
	}

	artifact, err := pki.LoadSignedArtifact[T](dbEntry.Config, h.verifier)
	if err != nil {
		panic(err)
	}

	return artifact, true
}

func (h *HostConfigHandler[T]) GetAll() map[string]*pki.SignedArtifact[T] {

	dbEntries, err := h.db.HostConfig.Query().WithDevice().Where(
		hostconfig.Type(h.configKey),
	).All(context.Background())

	if err != nil {
		panic(err)
	}

	artifacts := make(map[string]*pki.SignedArtifact[T], len(dbEntries))

	for _, dbEntry := range dbEntries {

		artifact, err := pki.LoadSignedArtifact[T](dbEntry.Config, h.verifier)
		if err != nil {
			panic(err)
		}

		artifacts[dbEntry.Edges.Device.PublicKey] = artifact
	}

	return artifacts
}

func (h *HostConfigHandler[T]) Size() int {
	return h.db.HostConfig.Query().Where(
		hostconfig.Type(h.configKey),
	).CountX(context.Background())
}

func (h *HostConfigHandler[T]) UpdateConfig(config *pki.SignedArtifact[T]) error {
	pubKey := config.Artifact().GetHost().Base64Encode()

	current, ok := h.Get(pubKey)
	if !ok {
		device, err := h.db.Device.Query().Where(
			device.PublicKey(pubKey),
		).Only(context.Background())

		if err != nil {
			return fmt.Errorf("error querying device: %w", err)
		}

		err = h.db.HostConfig.Create().
			SetType(h.configKey).
			SetConfig(config.Raw()).
			SetDevice(device).
			Exec(context.Background())

		if err != nil {
			return fmt.Errorf("error saving host config: %w", err)
		}
	} else {
		if current.Timestamp() >= config.Timestamp() {
			return fmt.Errorf("old config received")
		}

		err := h.db.HostConfig.Update().Where(
			hostconfig.Type(h.configKey),
			hostconfig.HasDeviceWith(device.PublicKey(pubKey)),
		).Exec(context.Background())

		if err != nil {
			return fmt.Errorf("error updating host config: %w", err)
		}
	}

	h.observerHandler.NotifyUpdate(pubKey, config)
	return nil
}
