package rmm

import (
	"context"
	"fmt"
	"log"
	"rahnit-rmm/config"
	"rahnit-rmm/ent/device"
	"rahnit-rmm/ent/hostconfig"
	"rahnit-rmm/pki"
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

	artifact, err := pki.LoadSignedArtifact[T](savedConfig.Config, verifier, hostConf)
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

func SaveHostConfigToDB[T HostConfig](hostConf pki.SignedArtifact[T]) error {
	db := config.DB()

	conf := hostConf.Artifact()

	host := conf.GetHost().Base64Encode()
	key := conf.GetConfigKey()

	db.HostConfig.Create().
		SetType(hostConf.GetConfigKey()).
		SetConfig(hostConf.Raw())
}
