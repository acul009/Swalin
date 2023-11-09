package pki

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/asn1"
	"encoding/json"
	"fmt"
	"io"
	"rahnit-rmm/util"
)

var ErrSignatureInvalid = SignatureVerificationError{
	Signature: []byte{},
	PublicKey: nil,
}

var ErrNotSigned = NotSignedError{}

type SignatureVerificationError struct {
	Signature []byte
	PublicKey *PublicKey
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

func MarshalAndSign(v any, c Credentials) ([]byte, error) {
	json, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	msg, err := packAndSign(json, c)
	if err != nil {
		return nil, fmt.Errorf("failed to package data: %w", err)
	}

	return msg, nil
}

func UnmarshalAndVerify(signedData []byte, v any, publicKey *PublicKey, checkRevocation bool) error {
	if len(signedData) == 0 {
		return fmt.Errorf("empty signed data")
	}

	msg, err := unpackAndVerify(signedData, publicKey, checkRevocation)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}

	err = json.Unmarshal(msg, v)
	if err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

func ReadAndUnmarshalAndVerify(reader io.Reader, v any, publicKey *PublicKey, checkRevocation bool) error {
	der, err := util.ReadSingleDer(reader)
	if err != nil {
		return fmt.Errorf("failed to read asn1 block: %w", err)
	}

	return UnmarshalAndVerify(der, v, publicKey, checkRevocation)
}

type PackedData struct {
	Data      []byte `asn1:"tag:0"`
	Signature []byte `asn1:"tag:1"`
}

func packAndSign(data []byte, c Credentials) ([]byte, error) {
	key, err := c.GetPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	signature, err := key.signBytes(data)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	d := PackedData{
		Data:      data,
		Signature: signature,
	}

	packed, err := asn1.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("failed to pack data to asn1: %w", err)
	}

	return packed, nil
}

func unpackAndVerify(packed []byte, publicKey *PublicKey, checkRevocation bool) ([]byte, error) {
	d, err := unpack(packed)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack data: %w", err)
	}

	err = publicKey.verifyBytes(d.Data, d.Signature, checkRevocation)
	if err != nil {
		return nil, fmt.Errorf("failed to verify signature: %w", err)
	}

	return d.Data, nil
}

func unpack(packed []byte) (*PackedData, error) {
	d := &PackedData{}

	rest, err := asn1.Unmarshal(packed, d)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}
	if len(rest) > 0 {
		return nil, fmt.Errorf("found rest after unmarshaling data: %v", rest)
	}

	return d, nil
}

func (p *PrivateKey) signBytes(data []byte) ([]byte, error) {

	hash := crypto.SHA512.New().Sum(data)

	signature, err := ecdsa.SignASN1(rand.Reader, p.ToEcdsa(), hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	return signature, nil
}

func (pub *PublicKey) verifyBytes(data []byte, signature []byte, checkRevocation bool) error {
	if pub == nil {
		return fmt.Errorf("public key cannot be nil")
	}

	hash := crypto.SHA512.New().Sum(data)

	if checkRevocation {
		err := RevocationManager.CheckPayload(data)
		if err != nil {
			return fmt.Errorf("failed revocation check: %w", err)
		}
	}

	ok := ecdsa.VerifyASN1(pub.ToEcdsa(), hash, signature)

	if !ok {
		return SignatureVerificationError{
			Signature: signature,
			PublicKey: pub,
		}
	}

	return nil
}
