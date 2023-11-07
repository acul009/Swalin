package rmm

import "rahnit-rmm/pki"

type hostConfig[T any] struct {
	pki.SignedArtifact[hostConfigPayload[T]]
}

type hostConfigPayload[T any] struct {
	Host   *pki.PublicKey
	Config T
}

func LoadHostConfig[T any](raw []byte) (*hostConfig[T], error) {
	artifact, err := pki.LoadSignedArtifact[hostConfigPayload[T]](raw)
	if err != nil {
		return nil, err
	}

	return &hostConfig[T]{*artifact}, nil
}

func CreateHostConfig[T any](host *pki.PublicKey, credentials pki.Credentials, config T) (*hostConfig[T], error) {
	payload := hostConfigPayload[T]{
		Host:   host,
		Config: config,
	}

	artifact, err := pki.NewSignedArtifact[hostConfigPayload[T]](credentials, payload)
	if err != nil {
		return nil, err
	}

	return &hostConfig[T]{*artifact}, nil
}
