package rpc

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"time"
)

const messageExpiration = 30

type RpcMessage[P any] struct {
	Timestamp int64
	Receiver  *ecdsa.PublicKey
	Nonce     Nonce
	Payload   P
}

type jsonRpcMessage[P any] struct {
	Timestamp int64
	Receiver  []byte
	Nonce     []byte
	Payload   P
}

func newRpcMessage[P any](receiver *ecdsa.PublicKey, payload P) (*RpcMessage[P], error) {
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

func (m *RpcMessage[P]) MarshalJSON() ([]byte, error) {
	pubKey, err := x509.MarshalPKIXPublicKey(m.Receiver)
	if err != nil {
		return nil, fmt.Errorf("error marshalling public key: %w", err)
	}
	return json.Marshal(&jsonRpcMessage[P]{
		Timestamp: m.Timestamp,
		Receiver:  pubKey,
		Nonce:     m.Nonce,
		Payload:   m.Payload,
	})
}

func (m *RpcMessage[P]) UnmarshalJSON(data []byte) error {
	var message jsonRpcMessage[P]
	err := json.Unmarshal(data, &message)
	if err != nil {
		return fmt.Errorf("error unmarshalling message: %w", err)
	}
	m.Timestamp = message.Timestamp
	pubKey, err := x509.ParsePKIXPublicKey(message.Receiver)
	if err != nil {
		return fmt.Errorf("error parsing public key: %w", err)
	}
	m.Receiver = pubKey.(*ecdsa.PublicKey)
	m.Nonce = message.Nonce
	m.Payload = message.Payload
	return nil
}

func (m *RpcMessage[P]) Verify(store *nonceStorage, receiver *ecdsa.PublicKey) error {
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
	if time.Now().Unix()-m.Timestamp > messageExpiration {
		return fmt.Errorf("message expired")
	}
	return nil
}

func (m *RpcMessage[P]) VerifyNonce(store *nonceStorage) error {
	if !store.CheckNonce(m.Nonce) {
		return fmt.Errorf("invalid nonce")
	}
	store.AddNonce(m.Nonce)
	return nil
}

func (m *RpcMessage[P]) VerifyReceiver(receiver *ecdsa.PublicKey) error {
	if !m.Receiver.Equal(receiver) {
		return fmt.Errorf("invalid receiver")
	}
	return nil
}
