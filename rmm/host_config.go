package rmm

import "rahnit-rmm/pki"

type HostConfig interface {
	pki.ArtifactPayload
	MayAccess(*pki.Certificate) bool
	GetHost() *pki.PublicKey
	GetConfigKey() string
}
