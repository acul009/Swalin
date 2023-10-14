package pki

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

func GetEncodedPublicKey(cert *x509.Certificate) (string, error) {
	key, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("public key is not of type *ecdsa.PublicKey")
	}
	return EncodePubToString(key)
}

func EncodePubToString(pub *ecdsa.PublicKey) (string, error) {
	encoded, err := EncodePubToBytes(pub)
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(encoded), nil
}

func EncodePubToBytes(pub *ecdsa.PublicKey) ([]byte, error) {
	key, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	return key, nil
}

func DecodePubFromString(pubString string) (*ecdsa.PublicKey, error) {
	raw, err := base64.RawStdEncoding.DecodeString(pubString)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	return DecodePubFromBytes(raw)
}

func DecodePubFromBytes(encodedPub []byte) (*ecdsa.PublicKey, error) {
	pub, err := x509.ParsePKIXPublicKey(encodedPub)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	pubTyped, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not of type *ecdsa.PublicKey")
	}

	return pubTyped, nil
}

func EncodeCertificate(cert *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
}

func DecodeCertificate(certPEM []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

func EncodePrivateKeyToPEM(key *ecdsa.PrivateKey) ([]byte, error) {
	privBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privBytes,
	})

	return keyPEM, nil
}
