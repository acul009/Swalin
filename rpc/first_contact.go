package rpc

import (
	"encoding/asn1"
	"fmt"
	"io"
	"log"

	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/util"
)

func exchangeKeys(session *RpcSession) error {
	err := sendMyKey(session)
	if err != nil {
		return fmt.Errorf("error sending my public key: %w", err)
	}

	err = receivePartnerKey(session)
	if err != nil {
		return fmt.Errorf("error receiving partner public key: %w", err)
	}

	return nil
}

type keyPayload struct {
	PubKey []byte
}

func receivePartnerKey(session *RpcSession) error {

	log.Printf("Receiving partner public key...")

	derMessage, err := util.ReadSingleDer(session)
	if err != nil {
		return fmt.Errorf("error reading public key: %w", err)
	}

	payload := &keyPayload{}
	_, err = asn1.Unmarshal(derMessage, payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	partnerKey, err := pki.PublicKeyFromBinary(payload.PubKey)
	if err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	session.partnerKey = partnerKey

	return nil
}

func sendMyKey(session *RpcSession) error {
	credentials := session.credentials

	packed, err := asn1.Marshal(keyPayload{
		PubKey: credentials.PublicKey().BinaryEncode(),
	})
	if err != nil {
		return fmt.Errorf("failed to pack data to asn1: %w", err)
	}

	n, err := session.Write(packed)
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}
	if n != len(packed) {
		return fmt.Errorf("error writing message: %w", io.ErrShortWrite)
	}

	return nil
}
