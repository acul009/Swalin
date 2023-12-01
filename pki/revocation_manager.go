package pki

import (
	"bytes"
	"crypto"
	"encoding/base64"
	"fmt"
	"log"
)

var ErrRevoked = &revokedError{}

type revokedError struct {
	Revocation *Revocation
}

func (e *revokedError) Error() string {
	return fmt.Sprintf("revoked: %v", e.Revocation)
}

func (e *revokedError) Is(target error) bool {
	_, ok := target.(*revokedError)
	return ok
}

var RevocationManager *revocationManager

type revocationManager struct {
	verifier Verifier
}

func InitRevocationManager(verifier Verifier) {

	RevocationManager = &revocationManager{
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

		err := r.checkRevokedHash(hash, hasher)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *revocationManager) checkRevokedHash(hash []byte, hasher crypto.Hash) error {
	baseHash := base64.StdEncoding.EncodeToString(hash)

	revocation, err := RevocationFromBinary(revModel.Revocation)
	if err != nil {
		errDangerous := fmt.Errorf("WARNING: failed to decode revocation: %w", err)
		log.Print(errDangerous)
		return fmt.Errorf("WARNING: failed to load revocation: %w", err)
	}

	revoked := revocation.Hasher == hasher && bytes.Equal(revocation.Hash, hash)
	if !revoked {
		log.Printf("WARNING: revocation for %x has broken index", hash)
	}

	if revoked {
		return &revokedError{
			Revocation: revocation,
		}
	}

	return nil
}
