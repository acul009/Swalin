package rmm

import (
	"github.com/rahn-it/svalin/pki"
)

type HostConfig interface {
	pki.ArtifactPayload
	MayAccess(*pki.Certificate) bool
	GetHost() *pki.PublicKey
	GetConfigKey() string
}
