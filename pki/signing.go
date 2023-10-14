package pki

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"rahnit-rmm/util"
)

var ErrSignatureInvalid = SignatureVerificationError{
	Signature: []byte{},
	PublicKey: nil,
}

var ErrNotSigned = NotSignedError{}

type SignatureVerificationError struct {
	Signature []byte
	PublicKey *ecdsa.PublicKey
}

func (e SignatureVerificationError) Error() string {
	return "failed to verify signature"
}

func (e SignatureVerificationError) Is(target error) bool {
	_, ok := target.(SignatureVerificationError)
	return ok
}

type NotSignedError struct {
}

func (e NotSignedError) Error() string {
	return "data is not signed"
}

func (e NotSignedError) Is(target error) bool {
	_, ok := target.(NotSignedError)
	return ok
}

var jsonDelimiter = []byte("|")

func MarshalAndSign(v any, key *ecdsa.PrivateKey, pub *ecdsa.PublicKey) ([]byte, error) {
	if key == nil {
		return nil, fmt.Errorf("private key cannot be nil")
	}

	if pub == nil {
		return nil, fmt.Errorf("public key cannot be nil")
	}

	json, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	signature, err := SignBytes(json, key)

	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	bsig := util.Base64Encode(signature)

	bpub, err := EncodePubToString(pub)
	if err != nil {
		return nil, fmt.Errorf("failed to encode public key: %w", err)
	}

	msg := json
	msg = append(msg, jsonDelimiter...)
	msg = append(msg, bsig...)
	msg = append(msg, jsonDelimiter...)
	msg = append(msg, bpub...)

	return msg, nil
}

func UnmarshalAndVerify(signedData []byte, v any) (*ecdsa.PublicKey, error) {
	if len(signedData) == 0 {
		return nil, fmt.Errorf("empty signed data")
	}

	split := bytes.SplitN(signedData, jsonDelimiter, 3)

	if len(split) == 1 {
		return nil, NotSignedError{}
	}
	if len(split) != 3 {
		return nil, fmt.Errorf("invalid signed data")
	}

	msg := split[0]
	bsig := split[1]
	bpub := split[2]

	pub, err := DecodePubFromString(string(bpub))
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	signature, err := util.Base64Decode(bsig)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	err = VerifyBytes(msg, signature, pub)
	if err != nil {
		return nil, fmt.Errorf("failed to verify signature: %w", err)
	}

	json.Unmarshal(msg, v)

	return pub, nil
}

func SignBytes(data []byte, key *ecdsa.PrivateKey) ([]byte, error) {
	if key == nil {
		return nil, fmt.Errorf("private key cannot be nil")
	}

	hash, err := HashBytes(data)
	if err != nil {
		return nil, fmt.Errorf("failed to hash data: %w", err)
	}

	signature, err := ecdsa.SignASN1(rand.Reader, key, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	return signature, nil
}

func VerifyBytes(data []byte, signature []byte, pub *ecdsa.PublicKey) error {
	if pub == nil {
		return fmt.Errorf("public key cannot be nil")
	}

	hash, err := HashBytes(data)
	if err != nil {
		return fmt.Errorf("failed to hash data: %w", err)
	}

	ok := ecdsa.VerifyASN1(pub, hash, signature)

	if !ok {
		return SignatureVerificationError{
			Signature: signature,
			PublicKey: pub,
		}
	}

	return nil
}

func HashBytes(data []byte) ([]byte, error) {
	hasher := crypto.SHA512.New()
	n, err := hasher.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to hash data: %w", err)
	}
	if n != len(data) {
		return nil, fmt.Errorf("failed to hash data: short write")
	}

	return hasher.Sum(nil), nil
}
