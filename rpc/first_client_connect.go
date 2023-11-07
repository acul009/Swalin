package rpc

import (
	"context"
	"fmt"
	"rahnit-rmm/pki"
	"rahnit-rmm/util"
	"time"

	"github.com/quic-go/quic-go"
)

func FirstClientConnect(addr string) (*RpcConnection, error) {
	tlsConf := getTlsTempClientConfig([]TlsConnectionProto{ProtoClientLogin, ProtoServerInit})

	quicConf := &quic.Config{
		KeepAlivePeriod: 30 * time.Second,
	}

	quicConn, err := quic.DialAddr(context.Background(), addr, tlsConf, quicConf)
	if err != nil {
		qErr, ok := err.(*quic.TransportError)
		if ok && uint8(qErr.ErrorCode) == 120 {
			return nil, fmt.Errorf("server not ready for login: %w", err)
		}
		return nil, fmt.Errorf("error creating QUIC connection: %w", err)
	}

	initNonceStorage = util.NewNonceStorage()

	protocol := quicConn.ConnectionState().TLS.NegotiatedProtocol

	conn := newRpcConnection(quicConn, nil, RpcRoleInit, initNonceStorage, nil, TlsConnectionProto(protocol), nil, pki.NewNilVerifier())

	return conn, nil
}
