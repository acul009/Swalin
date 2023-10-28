package rpc

import (
	"crypto/tls"
	"fmt"
	"rahnit-rmm/pki"
)

type TlsConnectionProto string

const (
	ProtoError       TlsConnectionProto = ""
	ProtoServerInit  TlsConnectionProto = "rahnit-rmm-server-init"
	ProtoRpc         TlsConnectionProto = "rahnit-rmm-rpc"
	ProtoClientLogin TlsConnectionProto = "rahnit-rmm-client-login"
)

func getTlsTempClientConfig(protos []TlsConnectionProto) *tls.Config {
	tlsProtos := make([]string, len(protos))

	for i, proto := range protos {
		tlsProtos[i] = string(proto)
	}

	return &tls.Config{
		// TODO: implement ACME certificate request and remove the InsecureSkipVerify option
		InsecureSkipVerify:   true,
		NextProtos:           tlsProtos,
		GetClientCertificate: nil,
	}
}

func getTlsClientConfig(proto TlsConnectionProto) *tls.Config {
	return &tls.Config{
		// TODO: implement ACME certificate request and remove the InsecureSkipVerify option
		InsecureSkipVerify: true,
		NextProtos:         []string{string(proto)},
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
}

func getTlsServerConfig(protos []TlsConnectionProto) (*tls.Config, error) {

	tlsCert, err := getServerCert()
	if err != nil {
		return nil, fmt.Errorf("error getting server cert: %w", err)
	}

	tlsProtos := make([]string, len(protos))

	for i, proto := range protos {
		tlsProtos[i] = string(proto)
	}

	return &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         tlsProtos,
		ClientAuth:         tls.RequestClientCert,
		Certificates:       []tls.Certificate{*tlsCert},
	}, nil
}
