package pki

import (
	"context"
	"crypto"
	"encoding/base64"
	"encoding/json"
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
	revModel, err := r.db.Revocation.Query().Where(revocation.HashEQ(base64.StdEncoding.EncodeToString(hash)), revocation.HasherEQ(uint64(hasher))).Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return false
		}
		log.Printf("WARNING: failed to query revocation: %v", err)
		return true
	}

	revocation := &Revocation{}
	err = json.Unmarshal([]byte(revModel.Revocation), revocation)
	if err != nil {
		log.Printf("WARNING: failed to unmarshal revocation: %v", err)
		return true
	}

	chain, err := r.verifier.VerifyPublicKey(revocation.Creator())
	if err != nil {
		log.Printf("WARNING: failed to verify revocation: %v", err)
		return false
	}

	if chain[0].Type() != CertTypeRoot && chain[0].Type() != CertTypeUser {
		log.Printf("WARNING: revocation not made by user or root: %v", err)
		return false
	}

	return true
}
