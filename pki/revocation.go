package pki

import (
	"crypto"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"rahnit-rmm/util"
)

type revocationPayload struct {
	Hash      []byte
	Hasher    crypto.Hash
	Timestamp int64
	Nonce     util.Nonce
	Chain     []*Certificate
}

type Revocation struct {
	PackedPayload []byte
	Signature     []byte
}

func RevocationFromBase64(data string) (*Revocation, error) {
	binary, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

}

func (r Revocation) Payload() (*revocationPayload, error) {
	payload := &revocationPayload{}
	err := json.Unmarshal(r.PackedPayload, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return payload, nil
}
