package pki

import (
	"encoding/asn1"
	"fmt"
	"log"
	"rahnit-rmm/util"
	"time"
)

type SignedBlob struct {
	creator   *Certificate
	timestamp int64
	nonce     util.Nonce
	payload   []byte
	raw       []byte
}

type blobDerHelper struct {
	Creator   []byte
	Timestamp int64
	Nonce     util.Nonce
	Payload   []byte
}

func NewSignedBlob(credentials *PermanentCredentials, payload []byte) (*SignedBlob, error) {
	creator, err := credentials.GetCertificate()
	if err != nil {
		return nil, fmt.Errorf("failed to get current cert: %w", err)
	}

	timeStamp := time.Now().Unix()

	nonce, err := util.NewNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	blobHelper := blobDerHelper{
		Creator:   creator.BinaryEncode(),
		Timestamp: timeStamp,
		Nonce:     nonce,
		Payload:   payload,
	}

	encodedBlob, err := asn1.Marshal(blobHelper)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal blob: %w", err)
	}

	signedPayload, err := packAndSign(encodedBlob, credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to sign blob: %w", err)
	}

	return &SignedBlob{
		creator:   creator,
		timestamp: timeStamp,
		nonce:     nonce,
		payload:   payload,
		raw:       signedPayload,
	}, nil
}

func (s *SignedBlob) Raw() []byte {
	return s.raw
}

func (s *SignedBlob) Payload() []byte {
	return s.payload
}

func (s *SignedBlob) Timestamp() int64 {
	return s.timestamp
}

func (s *SignedBlob) Creator() *Certificate {
	return s.creator
}

func LoadSignedBlob(raw []byte, verifier Verifier) (*SignedBlob, error) {
	packed, err := unpack(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack blob: %w", err)
	}

	blobHelper := &blobDerHelper{}

	rest, err := asn1.Unmarshal(packed.Data, blobHelper)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal blob: %w", err)
	}
	if len(rest) > 0 {
		return nil, fmt.Errorf("found rest after unmarshaling blob: %v", rest)
	}

	cert, err := CertificateFromBinary(blobHelper.Creator)
	if err != nil {
		return nil, fmt.Errorf("failed to parse blob creator: %w", err)
	}

	_, err = verifier.Verify(cert)
	if err != nil {
		errDangerous := fmt.Errorf("WARNING: failed to verify blob creator, server may be compromised: %w", err)
		log.Print(errDangerous)
		return nil, errDangerous
	}

	err = cert.GetPublicKey().verifyBytes(packed.Data, packed.Signature, true)
	if err != nil {
		return nil, fmt.Errorf("failed to verify signature: %w", err)
	}

	return &SignedBlob{
		creator:   cert,
		timestamp: blobHelper.Timestamp,
		nonce:     blobHelper.Nonce,
		payload:   blobHelper.Payload,
		raw:       raw,
	}, nil
}
