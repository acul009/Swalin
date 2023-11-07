package pki

import (
	"crypto"
	"encoding/asn1"
	"fmt"
	"rahnit-rmm/util"
)

type Revocation struct {
	Hash      []byte
	Hasher    crypto.Hash
	Timestamp int64
	Nonce     util.Nonce
	raw       []byte
}

type revocationDerHelper struct {
	payload []byte
	Chain   []*Certificate
}

func RevocationFromBinary(data []byte) (*Revocation, error) {

	helper := &revocationDerHelper{}

	_, err := asn1.Unmarshal(data, helper)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal revocation helper: %w", err)
	}

	_, err = helper.Chain[0].VerifyChain(nil, CreatePool(helper.Chain[1:]), false)
	if err != nil {
		return nil, fmt.Errorf("failed to verify chain: %w", err)
	}

	revocation := &Revocation{}
	_, err = asn1.Unmarshal(helper.payload, revocation)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal revocation: %w", err)
	}

	revocation.raw = data

	return revocation, nil
}
