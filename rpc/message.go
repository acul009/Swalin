package rpc

import (
	"crypto/ecdsa"
	"fmt"
	"time"
)

const messageExpiration = 30

type rpcMessage[P any] struct {
	timestamp int64
	receiver  *ecdsa.PublicKey
	nonce     Nonce
	payload   P
}

func newRpcMessage[P any](receiver *ecdsa.PublicKey, payload P) (*rpcMessage[P], error) {
	nonce, err := NewNonce()
	if err != nil {
		return nil, fmt.Errorf("error generating nonce: %v", err)
	}

	return &rpcMessage[P]{
		timestamp: time.Now().Unix(),
		receiver:  receiver,
		nonce:     nonce,
		payload:   payload,
	}, nil
}

func (m *rpcMessage[P]) Verify(store *nonceStorage, receiver *ecdsa.PublicKey) error {
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

func (m *rpcMessage[P]) VerifyTimestamp() error {
	if time.Now().Unix()-m.timestamp > messageExpiration {
		return fmt.Errorf("message expired")
	}
	return nil
}

func (m *rpcMessage[P]) VerifyNonce(store *nonceStorage) error {
	if !store.CheckNonce(m.nonce) {
		return fmt.Errorf("invalid nonce")
	}
	store.AddNonce(m.nonce)
	return nil
}

func (m *rpcMessage[P]) VerifyReceiver(receiver *ecdsa.PublicKey) error {
	if !m.receiver.Equal(receiver) {
		return fmt.Errorf("invalid receiver")
	}
	return nil
}
