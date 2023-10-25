package rpc

import (
	"fmt"
	"io"
	"log"
	"rahnit-rmm/pki"
)

func receivePartnerKey(session *RpcSession) error {

	log.Printf("Receiving partner public key...")

	var pubRoot *pki.PublicKey = nil
	sender, err := pki.ReadAndUnmarshalAndVerify(session, &pubRoot)
	if err != nil {
		return fmt.Errorf("error reading public key: %w", err)
	}

	if !sender.Equal(pubRoot) {
		return fmt.Errorf("partner public key does not match sender")
	}

	session.partner = pubRoot

	log.Printf("Received partner public key")

	return nil
}

func sendMyKey(session *RpcSession) error {
	pubKey, err := pki.GetCurrentPublicKey()
	if err != nil {
		return fmt.Errorf("error getting current public key: %w", err)
	}

	key, err := pki.GetCurrentKey()
	if err != nil {
		return fmt.Errorf("error getting current key: %w", err)
	}

	payload, err := pki.MarshalAndSign(pubKey, key, pubKey)
	if err != nil {
		return fmt.Errorf("error marshalling message: %w", err)
	}

	n, err := session.Write(payload)
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}
	if n != len(payload) {
		return fmt.Errorf("error writing message: %w", io.ErrShortWrite)
	}

	log.Printf("Sent my public key")

	return nil
}
