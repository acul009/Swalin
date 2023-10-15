package rpc

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"rahnit-rmm/pki"
	"time"
)

const messageExpiration = 30

type RpcMessage[P any] struct {
	Timestamp int64
	Receiver  []byte
	Nonce     Nonce
	Payload   P
}

func newRpcMessage[P any](receiver *ecdsa.PublicKey, payload P) (*RpcMessage[P], error) {
	nonce, err := NewNonce()
	if err != nil {
		return nil, fmt.Errorf("error generating nonce: %w", err)
	}

	receiverBytes, err := pki.EncodePubToBytes(receiver)
	if err != nil {
		return nil, fmt.Errorf("error encoding receiver: %w", err)
	}

	return &RpcMessage[P]{
		Timestamp: time.Now().Unix(),
		Receiver:  receiverBytes,
		Nonce:     nonce,
		Payload:   payload,
	}, nil
}

func (m *RpcMessage[P]) Verify(store *nonceStorage, receiver *ecdsa.PublicKey) error {
	log.Printf("Verifying message: %+v", m)

	if err := m.VerifyTimestamp(); err != nil {
		return err
	}

	if err := m.VerifyNonce(store); err != nil {
		return err
	}

	if err := m.VerifyReceiver(receiver); err != nil {
		return err
	}

	return nil
}

func (m *RpcMessage[P]) VerifyTimestamp() error {
	diff := time.Now().Unix() - m.Timestamp
	if diff > messageExpiration {
		return fmt.Errorf("message expired, singed at %d, now is %d, off by %d seconds", m.Timestamp, time.Now().Unix(), messageExpiration)
	}
	return nil
}

func (m *RpcMessage[P]) VerifyNonce(store *nonceStorage) error {
	if !store.CheckNonce(m.Nonce) {
		return fmt.Errorf("nonce has already been used, possible replay attack")
	}
	store.AddNonce(m.Nonce)
	return nil
}

func (m *RpcMessage[P]) VerifyReceiver(receiver *ecdsa.PublicKey) error {
	actualReceiver, err := pki.DecodePubFromBytes(m.Receiver)
	if err != nil {
		return fmt.Errorf("error decoding receiver: %w", err)
	}

	if !receiver.Equal(actualReceiver) {
		return fmt.Errorf("message was meant for someone else, possible replay attack")
	}
	return nil
}
