package rmm

import (
	"fmt"
	"log"

	"github.com/rahn-it/svalin/pki"
)

type HostConfig interface {
	pki.ArtifactPayload
	MayAccess(*pki.Certificate) bool
	GetHost() *pki.PublicKey
	GetConfigKey() string
}

func LoadHostConfigFromDB[T HostConfig](host *pki.PublicKey, verifier pki.Verifier) (*pki.SignedArtifact[T], error) {

	var hostConf T

	configKey := hostConf.GetConfigKey()

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

	conf := hostConf.Artifact()

	host := conf.GetHost().Base64Encode()

	key := conf.GetConfigKey()

	return nil
}
