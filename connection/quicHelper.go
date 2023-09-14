package connection

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"rahnit-rmm/rpc"

	"github.com/quic-go/quic-go"
)

func CreateClient(ctx context.Context, addr string) (*rpc.NodeConnection, error) {
	tlsConfig := generateClientTLSConfig()
	quicConfig := generateQuicConfig()
	conn, err := quic.DialAddr(ctx, addr, tlsConfig, quicConfig)
	if err != nil {
		return nil, err
	}
	return rpc.NewNodeConnection(conn, nil), nil
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
	}
}

func generateClientTLSConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}
}

func generateQuicConfig() *quic.Config {
	return &quic.Config{}
}
