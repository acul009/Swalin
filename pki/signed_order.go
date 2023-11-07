package pki

import "rahnit-rmm/util"

type signedOrder struct {
	signature []byte
	timestamp int64
	nonce     util.Nonce
	data      []byte
}
