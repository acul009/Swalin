package pki

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

var ErrSignatureInvalid = SignatureVerificationError{
	Signature: []byte{},
	PublicKey: nil,
}

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
		return nil, fmt.Errorf("failed to marshal data: %v", err)
	}

	signature, err := SignBytes(json, key)

	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %v", err)
	}

	bsig := Base64Encode(signature)

	ecdsaPublicKeyBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %v", err)
	}
	bpub := Base64Encode(ecdsaPublicKeyBytes)

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
	if len(split) != 3 {
		return nil, fmt.Errorf("invalid signed data")
	}

	msg := split[0]
	bsig := split[1]
	bpub := split[2]

	ecdsaPublicKeyBytes, err := Base64Decode(bpub)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %v", err)
	}

	ecdsaPublicKey, err := x509.ParsePKIXPublicKey(ecdsaPublicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	signature, err := Base64Decode(bsig)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %v", err)
	}

	pub, ok := ecdsaPublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to parse public key: invalid type")
	}

	err = VerifyBytes(msg, signature, pub)
	if err != nil {
		return nil, fmt.Errorf("failed to verify signature: %v", err)
	}

	json.Unmarshal(msg, v)

	return pub, nil
}

func Base64Encode(data []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(data))
}

func Base64Decode(data []byte) ([]byte, error) {
	return base64.StdEncoding.DecodeString(string(data))
}

func SignBytes(data []byte, key *ecdsa.PrivateKey) ([]byte, error) {
	if key == nil {
		return nil, fmt.Errorf("private key cannot be nil")
	}

	hash, err := HashBytes(data)

	signature, err := ecdsa.SignASN1(rand.Reader, key, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %v", err)
	}

	return signature, nil
}

func VerifyBytes(data []byte, signature []byte, pub *ecdsa.PublicKey) error {
	if pub == nil {
		return fmt.Errorf("public key cannot be nil")
	}

	hash, err := HashBytes(data)
	if err != nil {
		return fmt.Errorf("failed to hash data: %v", err)
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
		return nil, fmt.Errorf("failed to hash data: %v", err)
	}
	if n != len(data) {
		return nil, fmt.Errorf("failed to hash data: short write")
	}

	return hasher.Sum(nil), nil
}
