package rpc

import (
	"crypto/ecdsa"
	"fmt"
	"time"
)

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
