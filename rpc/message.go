package rpc

import (
	"fmt"
	"rahnit-rmm/pki"
	"time"
)

const messageExpiration = 30

type RpcMessage[P any] struct {
	Timestamp int64
	Receiver  *pki.PublicKey
	Nonce     Nonce
	Payload   P
}

func newRpcMessage[P any](receiver *pki.PublicKey, payload P) (*RpcMessage[P], error) {
	nonce, err := NewNonce()
	if err != nil {
		return nil, fmt.Errorf("error generating nonce: %w", err)
	}

	return &RpcMessage[P]{
		Timestamp: time.Now().Unix(),
		Receiver:  receiver,
		Nonce:     nonce,
		Payload:   payload,
	}, nil
}

func (m *RpcMessage[P]) Verify(store *nonceStorage, receiver *pki.PublicKey) error {

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

func (m *RpcMessage[P]) VerifyReceiver(receiver *pki.PublicKey) error {
	actualReceiver := m.Receiver

	if !receiver.Equal(actualReceiver) {
		return fmt.Errorf("message was meant for someone else, possible replay attack")
	}
	return nil
}
