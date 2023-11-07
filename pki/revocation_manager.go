package pki

import (
	"bytes"
	"context"
	"crypto"
	"encoding/base64"
	"fmt"
	"log"
	"rahnit-rmm/config"
	"rahnit-rmm/ent"
	"rahnit-rmm/ent/revocation"
)

var RevocationManager *revocationManager

type revocationManager struct {
	db       *ent.Client
	verifier Verifier
}

func InitRevocationManager(verifier Verifier) {
	db := config.DB()

	RevocationManager = &revocationManager{
		db:       db,
		verifier: verifier,
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
	baseHash := base64.StdEncoding.EncodeToString(hash)

	revModel, err := r.db.Revocation.Query().Where(revocation.HashEQ(baseHash), revocation.HasherEQ(uint64(hasher))).Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return false
		}
		log.Printf("WARNING: failed to query revocation: %v", err)
		return true
	}

	revocation, err := RevocationFromBinary(revModel.Revocation)
	if err != nil {
		log.Printf("WARNING: failed to load revocation: %v", err)
		return false
	}

	revoked := revocation.Hasher == hasher && bytes.Equal(revocation.Hash, hash)
	if !revoked {
		log.Printf("WARNING: revocation for %x has broken index", hash)
	}

	return revoked
}
