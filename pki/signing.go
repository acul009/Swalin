package pki

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/asn1"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
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

	msg, err := packAndSign(json, key, pub)
	if err != nil {
		return nil, fmt.Errorf("failed to package data: %w", err)
	}

	return msg, nil
}

func UnmarshalAndVerify(signedData []byte, v any) (*ecdsa.PublicKey, error) {
	if len(signedData) == 0 {
		return nil, fmt.Errorf("empty signed data")
	}

	msg, pub, err := unpackAndVerify(signedData)
	if err != nil {
		return nil, fmt.Errorf("failed to verify signature: %w", err)
	}

	json.Unmarshal(msg, v)

	return pub, nil
}

func ReadAndUnmarshalAndVerify(reader io.Reader, v any) (*ecdsa.PublicKey, error) {
	der, err := ReadSingleDer(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read asn1 block: %w", err)
	}

	return UnmarshalAndVerify(der, v)
}

func ReadSingleDer(reader io.Reader) ([]byte, error) {
	derStart := make([]byte, 2)
	_, err := io.ReadFull(reader, derStart)
	if err != nil {
		return nil, fmt.Errorf("failed to first two asn1 bytes: %w", err)
	}

	isMultiByteLength := derStart[1]&0b1000_0000 != 0
	firstByteValue := derStart[1] & 0b0111_1111
	var lengthBytes []byte
	if isMultiByteLength {
		lengthBytes = make([]byte, uint(firstByteValue))
		_, err := io.ReadFull(reader, lengthBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to read extended asn1 length: %w", err)
		}
	} else {
		lengthBytes = []byte{firstByteValue}
	}

	length := &big.Int{}
	length.SetBytes(lengthBytes)

	derBody := make([]byte, length.Int64())
	_, err = io.ReadFull(reader, derBody)
	if err != nil {
		return nil, fmt.Errorf("failed to read asn1 body: %w", err)
	}

	toJoin := [][]byte{
		derStart,
	}

	if isMultiByteLength {
		toJoin = append(toJoin, lengthBytes)
	}
	toJoin = append(toJoin, derBody)

	return bytes.Join(toJoin, []byte{}), nil
}

type PackedData struct {
	Data      []byte `asn1:"tag:0"`
	Signature []byte `asn1:"tag:1"`
	PublicKey []byte `asn1:"tag:2"`
}

func packAndSign(data []byte, key *ecdsa.PrivateKey, pub *ecdsa.PublicKey) ([]byte, error) {

	signature, err := signBytes(data, key)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	pubData, err := EncodePubToBytes(pub)
	if err != nil {
		return nil, fmt.Errorf("failed to encode public key: %w", err)
	}

	d := PackedData{
		Data:      data,
		Signature: signature,
		PublicKey: pubData,
	}

	packed, err := asn1.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("failed to pack data to asn1: %w", err)
	}

	return packed, nil
}

func unpackAndVerify(packed []byte) ([]byte, *ecdsa.PublicKey, error) {
	d := &PackedData{}

	rest, err := asn1.Unmarshal(packed, d)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}
	if len(rest) > 0 {
		return nil, nil, fmt.Errorf("found rest after unmarshaling data: %v", rest)
	}

	pub, err := DecodePubFromBytes(d.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	err = verifyBytes(d.Data, d.Signature, pub)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to verify signature: %w", err)
	}

	return d.Data, pub, nil
}

func signBytes(data []byte, key *ecdsa.PrivateKey) ([]byte, error) {
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

func verifyBytes(data []byte, signature []byte, pub *ecdsa.PublicKey) error {
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
