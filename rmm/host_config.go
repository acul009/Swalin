package rmm

import (
	"context"
	"fmt"
	"log"

	"github.com/rahn-it/svalin/config"
	"github.com/rahn-it/svalin/ent/device"
	"github.com/rahn-it/svalin/ent/hostconfig"
	"github.com/rahn-it/svalin/pki"
)

type HostConfig interface {
	pki.ArtifactPayload
	MayAccess(*pki.Certificate) bool
	GetHost() *pki.PublicKey
	GetConfigKey() string
}

func LoadHostConfigFromDB[T HostConfig](host *pki.PublicKey, verifier pki.Verifier) (*pki.SignedArtifact[T], error) {
	db := config.DB()

	var hostConf T

	configKey := hostConf.GetConfigKey()

	savedConfig, err := db.HostConfig.Query().Where(hostconfig.Type(configKey), hostconfig.HasDeviceWith(device.PublicKey(host.Base64Encode()))).Only(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error querying host config: %w", err)
	}

	artifact, err := pki.LoadSignedArtifact[T](savedConfig.Config, verifier)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if !hostConf.GetHost().Equal(host) {
		err := fmt.Errorf("damaged index for host config")
		log.Print(err)
		return nil, err
	}

	return artifact, nil
}

func SaveHostConfigToDB[T HostConfig](hostConf *pki.SignedArtifact[T]) error {
	db := config.DB()

	conf := hostConf.Artifact()

	host := conf.GetHost().Base64Encode()

	dev, err := db.Device.Query().Where(device.PublicKey(host)).Only(context.Background())
	if err != nil {
		return fmt.Errorf("error querying device: %w", err)
	}

	key := conf.GetConfigKey()

	_, err = db.HostConfig.Create().
		SetType(key).
		SetConfig(hostConf.Raw()).
		SetDevice(dev).Save(context.Background())
	if err != nil {
		return fmt.Errorf("error saving host config: %w", err)
	}

	return nil
}
