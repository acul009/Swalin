package pki

import (
	"crypto"
	"fmt"
	"rahnit-rmm/config"
	"rahnit-rmm/ent"
)

var RevocationManager = newRevocationManager()

type revocationManager struct {
	db *ent.Client
}

func newRevocationManager() *revocationManager {
	db := config.DB()

	return &revocationManager{
		db: db,
	}
}

func (r *revocationManager) getHashers() []crypto.Hash {
	return []crypto.Hash{
		crypto.SHA512,
	}
}

func (r *revocationManager) CheckPayload(payload []byte) error {
	for _, hasher := range r.getHashers() {
		hash := hasher.New().Sum(payload)

		if r.isRevokedHash(hash, hasher) {
			return fmt.Errorf("hash %x is revoked", hash)
		}
	}

	return nil
}

func (r *revocationManager) isRevokedHash(hash []byte, hasher crypto.Hash) bool {
	// TODO: check the DB for this hash
	return false
}
