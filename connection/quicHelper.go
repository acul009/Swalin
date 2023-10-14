package connection

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"

	"github.com/quic-go/quic-go"
)

func CreateClient(ctx context.Context, addr string) (quic.Connection, error) {
	tlsConfig := generateClientTLSConfig()
	quicConfig := generateQuicConfig()
	conn, err := quic.DialAddr(ctx, addr, tlsConfig, quicConfig)

	if err != nil {
		return nil, fmt.Errorf("error creating QUIC client: %w", err)
	}

	return conn, nil
}

func CreateServer(addr string) (*quic.Listener, error) {
	tlsConfig := generateServerTLSConfig()
	quicConfig := generateQuicConfig()
	return quic.ListenAddr(addr, tlsConfig, quicConfig)
}

func generateServerTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo-example"},
		ClientAuth:   tls.RequireAnyClientCert,
		RootCAs:      x509.NewCertPool(),
	}
}

func generateClientTLSConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
		ClientAuth:         tls.VerifyClientCertIfGiven,
	}
}

func generateQuicConfig() *quic.Config {
	return &quic.Config{}
}
