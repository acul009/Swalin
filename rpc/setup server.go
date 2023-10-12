package rpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"rahnit-rmm/pki"

	"github.com/quic-go/quic-go"
)

func SetupServer(addr string) error {
	tlsCert, err := getServerCert()
	if err != nil {
		return fmt.Errorf("error getting server cert: %v", err)
	}

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
		ClientAuth:         tls.RequireAnyClientCert,
		Certificates:       []tls.Certificate{*tlsCert},
	}

	quicConf := &quic.Config{}
	listener, err := quic.ListenAddr(addr, tlsConf, quicConf)
	if err != nil {
		return fmt.Errorf("error creating QUIC server: %v", err)
	}

	initNonceStorage = NewNonceStorage()

	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			err := fmt.Errorf("error accepting QUIC connection: %v", err)
			log.Println(err)
		}

		err = acceptServerInitialization(conn)
		if err != nil {
			err := fmt.Errorf("error initializing server: %v", err)
			log.Println(err)
		} else {
			// no error, initialization was successful
			return nil
		}
	}
}

var initNonceStorage *nonceStorage = nil

type serverInitRequest struct {
	PubKey string
}

type serverInitResponse struct {
	RootCa     string
	ServerCert string
}

func acceptServerInitialization(quicConn quic.Connection) error {
	conn := NewRpcConnection(quicConn, nil, RpcRoleInit, initNonceStorage)

	session, err := conn.AcceptSession(context.Background())
	if err != nil {
		return fmt.Errorf("error accepting QUIC stream: %v", err)
	}

	pubKey, err := pki.GetCurrentPublicKey()
	if err != nil {
		return fmt.Errorf("error getting current public key: %v", err)
	}

	pubMarshalled, err := pki.EncodePubToString(pubKey)
	if err != nil {
		return fmt.Errorf("error marshalling public key: %v", err)
	}

	initRequest := &serverInitRequest{
		PubKey: pubMarshalled,
	}

	err = WriteMessage[*serverInitRequest](session, initRequest)
	if err != nil {
		return fmt.Errorf("error writing message: %v", err)
	}

	response := &serverInitResponse{}
	sender, err := readMessageFromUnknown[*serverInitResponse](session, response)
	if err != nil {
		return fmt.Errorf("error reading message: %v", err)
	}

	rootCert, err := pki.DecodeCertificate([]byte(response.RootCa))
	if err != nil {
		return fmt.Errorf("error decoding root certificate: %v", err)
	}

	if !sender.Equal(rootCert.PublicKey) {
		return fmt.Errorf("root certificate does not match sender")
	}

	serverCert, err := pki.DecodeCertificate([]byte(response.ServerCert))
	if err != nil {
		return fmt.Errorf("error decoding server certificate: %v", err)
	}

	err = pki.SaveCurrentCert(serverCert)
	if err != nil {
		return fmt.Errorf("error saving server certificate: %v", err)
	}

	err = pki.SaveRootCert(rootCert)
	if err != nil {
		return fmt.Errorf("error saving root certificate: %v", err)
	}

	return nil
}
