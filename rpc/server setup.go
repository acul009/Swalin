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
	ok, err := pki.CurrentAvailable()
	if err != nil {
		return fmt.Errorf("failed to check if current cert exists: %w", err)
	}

	if ok {
		// Server already initialized
		return nil
	}

	tlsCert, err := getServerCert()
	if err != nil {
		return fmt.Errorf("error getting server cert: %w", err)
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
		return fmt.Errorf("error creating QUIC server: %w", err)
	}

	initNonceStorage = NewNonceStorage()

	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			err := fmt.Errorf("error accepting QUIC connection: %w", err)
			log.Println(err)
		}

		err = acceptServerInitialization(conn)
		if err != nil {
			err := fmt.Errorf("error initializing server: %w", err)
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
		return fmt.Errorf("error accepting QUIC stream: %w", err)
	}

	pubKey, err := pki.GetCurrentPublicKey()
	if err != nil {
		return fmt.Errorf("error getting current public key: %w", err)
	}

	pubMarshalled, err := pki.EncodePubToString(pubKey)
	if err != nil {
		return fmt.Errorf("error marshalling public key: %w", err)
	}

	initRequest := &serverInitRequest{
		PubKey: pubMarshalled,
	}

	err = WriteMessage[*serverInitRequest](session, initRequest)
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}

	response := &serverInitResponse{}
	sender, err := readMessageFromUnknown[*serverInitResponse](session, response)
	if err != nil {
		return fmt.Errorf("error reading message: %w", err)
	}

	rootCert, err := pki.DecodeCertificate([]byte(response.RootCa))
	if err != nil {
		return fmt.Errorf("error decoding root certificate: %w", err)
	}

	if !sender.Equal(rootCert.PublicKey) {
		return fmt.Errorf("root certificate does not match sender")
	}

	serverCert, err := pki.DecodeCertificate([]byte(response.ServerCert))
	if err != nil {
		return fmt.Errorf("error decoding server certificate: %w", err)
	}

	err = pki.SaveCurrentCert(serverCert)
	if err != nil {
		return fmt.Errorf("error saving server certificate: %w", err)
	}

	err = pki.SaveRootCert(rootCert)
	if err != nil {
		return fmt.Errorf("error saving root certificate: %w", err)
	}

	session.Close()
	conn.Close(200, "done")

	return nil
}

func SetupServer(addr string, rootPassword []byte, nameForServer string) error {
	err := pki.UnlockAsRoot(rootPassword)
	if err != nil {
		return fmt.Errorf("error unlocking root cert: %w", err)
	}

	tlsConf := &tls.Config{
		// TODO: implement ACME certificate request and remove the InsecureSkipVerify option
		InsecureSkipVerify: true,
		NextProtos:         []string{serverInitProtocol},
		GetClientCertificate: func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			tlsCert, err := pki.GetCurrentTlsCert()
			if err != nil {
				return nil, fmt.Errorf("error getting current certificate: %w", err)
			}

			err = info.SupportsCertificate(tlsCert)
			if err != nil {
				return nil, fmt.Errorf("error checking certificate: %w", err)
			}
			return tlsCert, nil
		},
	}

	quicConf := &quic.Config{}

	quicConn, err := quic.DialAddr(context.Background(), addr, tlsConf, quicConf)
	if err != nil {
		return fmt.Errorf("error creating QUIC connection: %w", err)
	}

	initNonceStorage = NewNonceStorage()

	conn := NewRpcConnection(quicConn, nil, RpcRoleInit, initNonceStorage)

	log.Printf("Connection opened to %s\n", addr)

	session, err := conn.OpenSession(context.Background())
	if err != nil {
		return fmt.Errorf("error opening session: %w", err)
	}

	log.Printf("Session opened")

	req := &serverInitRequest{}

	sender, err := readMessageFromUnknown[*serverInitRequest](session, req)

	serverPubKey, err := pki.DecodePubFromString(req.PubKey)
	if err != nil {
		return fmt.Errorf("error decoding server public key: %w", err)
	}

	if !sender.Equal(serverPubKey) {
		return fmt.Errorf("server public key does not match sender")
	}

	log.Printf("Received request with pubkey: %s\n", req.PubKey)

	serverHostCert, err := pki.CreateServerCertWithCurrent(nameForServer, serverPubKey)
	if err != nil {
		return fmt.Errorf("error creating server certificate: %w", err)
	}

	log.Printf("Created server certificate:\n%s\n\n", string(pki.EncodeCertificate(serverHostCert)))

	rootCert, err := pki.GetRootCert()
	if err != nil {
		return fmt.Errorf("error getting root certificate: %w", err)
	}

	response := &serverInitResponse{
		RootCa:     string(pki.EncodeCertificate(rootCert)),
		ServerCert: string(pki.EncodeCertificate(serverHostCert)),
	}

	err = WriteMessage[*serverInitResponse](session, response)
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}

	session.Close()
	conn.Close(200, "done")

	return nil
}
