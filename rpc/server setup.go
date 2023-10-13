package rpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"rahnit-rmm/pki"

	"github.com/quic-go/quic-go"
)

const serverInitProtocol = "rahnit-rmm-server-init"

func WaitForServerSetup(listenAddr string) error {
	tlsCert, err := getServerCert()
	if err != nil {
		return fmt.Errorf("error getting server cert: %v", err)
	}

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{serverInitProtocol},
		ClientAuth:         tls.RequireAnyClientCert,
		Certificates:       []tls.Certificate{*tlsCert},
	}

	quicConf := &quic.Config{}
	listener, err := quic.ListenAddr(listenAddr, tlsConf, quicConf)
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

func SetupServer(addr string, rootPassword []byte) error {
	err := pki.UnlockAsRoot(rootPassword)
	if err != nil {
		return fmt.Errorf("error unlocking root cert: %v", err)
	}

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{serverInitProtocol},
		GetClientCertificate: func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			tlsCert, err := pki.GetCurrentTlsCert()
			if err != nil {
				return nil, err
			}

			err = info.SupportsCertificate(tlsCert)
			if err != nil {
				return nil, err
			}
			return tlsCert, nil
		},
	}

	quicConf := &quic.Config{}

	quicConn, err := quic.DialAddr(context.Background(), addr, tlsConf, quicConf)
	if err != nil {
		return fmt.Errorf("error creating QUIC connection: %v", err)
	}

	initNonceStorage = NewNonceStorage()

	conn := NewRpcConnection(quicConn, nil, RpcRoleInit, initNonceStorage)

	session, err := conn.OpenSession(context.Background())
	if err != nil {
		return fmt.Errorf("error opening session: %v", err)
	}

	req := &serverInitRequest{}

	sender, err := readMessageFromUnknown[*serverInitRequest](session, req)

	serverPubKey, err := pki.DecodePubFromString(req.PubKey)
	if err != nil {
		return fmt.Errorf("error decoding server public key: %v", err)
	}

	if !sender.Equal(serverPubKey) {
		return fmt.Errorf("server public key does not match sender")
	}

}
