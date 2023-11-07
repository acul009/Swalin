package pki

import "crypto"

type Revocation struct {
	SignedArtifact *SignedArtifact[revocationPayload]
}

type revocationPayload struct {
	Hash   []byte
	Hasher crypto.Hash
}
