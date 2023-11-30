package rpc

import (
	"context"
	"fmt"
	"github.com/rahn-it/svalin/config"
	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/util"
	"log"

	"github.com/quic-go/quic-go"
)

func WaitForServerSetup(listenAddr string) (*pki.PermanentCredentials, error) {

	credentials, err := pki.GenerateCredentials()
	if err != nil {
		return nil, fmt.Errorf("error generating temp credentials: %w", err)
	}

	tlsConf, err := getTlsServerConfig([]TlsConnectionProto{ProtoServerInit})
	if err != nil {
		return nil, fmt.Errorf("error getting server tls config: %w", err)
	}

	quicConf := &quic.Config{}
	listener, err := quic.ListenAddr(listenAddr, tlsConf, quicConf)
	if err != nil {
		return nil, fmt.Errorf("error creating QUIC server: %w", err)
	}

	initNonceStorage = util.NewNonceStorage()

	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			err := fmt.Errorf("error accepting QUIC connection: %w", err)
			log.Println(err)
			continue
		}

		log.Printf("Connection accepted")

		rpcCredentials, err := acceptServerInitialization(conn, credentials)
		if err != nil {
			err := fmt.Errorf("error initializing server: %w", err)
			log.Println(err)
		} else {
			// no error, initialization was successful
			err = listener.Close()
			if err != nil {
				return nil, fmt.Errorf("error closing listener: %w", err)
			}
			return rpcCredentials, nil
		}
	}
}

var initNonceStorage *util.NonceStorage = nil

type serverInitRequest struct {
	ServerPubKey *pki.PublicKey
}

type serverInitResponse struct {
	RootCa     *pki.Certificate
	ServerCert *pki.Certificate
}

func acceptServerInitialization(quicConn quic.Connection, credentials *pki.TempCredentials) (*pki.PermanentCredentials, error) {
	conn := newRpcConnection(quicConn, nil, RpcRoleInit, initNonceStorage, nil, ProtoServerInit, credentials, pki.NewNilVerifier())

	log.Printf("Opening init QUIC stream...")

	session, err := conn.AcceptSession(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error opening QUIC stream: %w", err)
	}

	err = session.mutateState(RpcSessionCreated, RpcSessionOpen)
	if err != nil {
		return nil, fmt.Errorf("error mutating session state: %w", err)
	}

	log.Printf("Session opened, reading public key...")

	err = exchangeKeys(session)
	if err != nil {
		return nil, fmt.Errorf("error exchanging keys: %w", err)
	}

	log.Printf("preparing init request...")

	pubMe, err := credentials.GetPublicKey()
	if err != nil {
		return nil, fmt.Errorf("error getting current public key: %w", err)
	}

	initRequest := &serverInitRequest{
		ServerPubKey: pubMe,
	}

	log.Printf("Sending init request...")

	err = WriteMessage[*serverInitRequest](session, initRequest)
	if err != nil {
		return nil, fmt.Errorf("error writing message: %w", err)
	}

	log.Printf("Awaiting init response...")

	response := &serverInitResponse{}
	err = ReadMessage[*serverInitResponse](session, response)
	if err != nil {
		return nil, fmt.Errorf("error reading message: %w", err)
	}

	log.Printf("Init response received")

	rootCert := response.RootCa

	if !session.partnerKey.Equal(rootCert.GetPublicKey()) {
		return nil, fmt.Errorf("root public key does not match certificate")
	}

	serverCert := response.ServerCert

	rpcCredentials, err := credentials.UpgradeToHostCredentials(serverCert)
	if err != nil {
		return nil, fmt.Errorf("error upgrading credentials: %w", err)
	}

	err = pki.Root.Set(rootCert)
	if err != nil {
		return nil, fmt.Errorf("error saving root certificate: %w", err)
	}

	session.Close()
	conn.Close(200, "done")

	return rpcCredentials, nil
}

func SetupServer(conn *RpcConnection, rootCredentials *pki.PermanentCredentials, nameForServer string) error {
	conn.credentials = rootCredentials

	session, err := conn.OpenSession(context.Background())
	if err != nil {
		return fmt.Errorf("error opening QUIC stream: %w", err)
	}

	err = session.mutateState(RpcSessionCreated, RpcSessionOpen)
	if err != nil {
		return fmt.Errorf("error mutating session state: %w", err)
	}

	log.Printf("Session opened, sending public key")

	err = exchangeKeys(session)
	if err != nil {
		return fmt.Errorf("error exchanging keys: %w", err)
	}

	log.Printf("reading initialization request^...")

	req := &serverInitRequest{}

	err = ReadMessage[*serverInitRequest](session, req)
	if err != nil {
		return fmt.Errorf("error reading message: %w", err)
	}

	if !session.partnerKey.Equal(req.ServerPubKey) {
		return fmt.Errorf("server public key does not match sender")
	}

	session.partnerKey = req.ServerPubKey

	log.Printf("Received request with pubkey: %s\n", req.ServerPubKey)

	serverHostCert, err := pki.CreateServerCert(nameForServer, req.ServerPubKey, rootCredentials)
	if err != nil {
		return fmt.Errorf("error creating server certificate: %w", err)
	}

	log.Printf("Created server certificate:\n%s\n\n", string(serverHostCert.PemEncode()))

	rootCert, err := pki.Root.Get()
	if err != nil {
		return fmt.Errorf("error getting root certificate: %w", err)
	}

	response := &serverInitResponse{
		RootCa:     rootCert,
		ServerCert: serverHostCert,
	}

	err = WriteMessage[*serverInitResponse](session, response)
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}

	session.Close()
	conn.Close(200, "done")

	config.V().Set("upstream.address", conn.connection.RemoteAddr().String())
	err = config.Save()
	if err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	err = pki.Upstream.Set(serverHostCert)
	if err != nil {
		return fmt.Errorf("error saving upstream certificate: %w", err)
	}

	return nil
}
