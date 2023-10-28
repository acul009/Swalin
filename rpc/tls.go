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

func getTlsClientConfig(proto TlsConnectionProto, credentials pki.Credentials) *tls.Config {
	var certGetter func(*tls.CertificateRequestInfo) (*tls.Certificate, error) = nil

	tlsCredentials, ok := credentials.(interface {
		GetTlsCert() (*tls.Certificate, error)
	})

	if ok {
		certGetter = func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			tlsCert, err := tlsCredentials.GetTlsCert()
			if err != nil {
				return nil, fmt.Errorf("error getting current certificate: %w", err)
			}

			err = info.SupportsCertificate(tlsCert)
			if err != nil {
				return nil, fmt.Errorf("error checking certificate: %w", err)
			}
			return tlsCert, nil
		}
	}

	return &tls.Config{
		// TODO: implement ACME certificate request and remove the InsecureSkipVerify option
		InsecureSkipVerify:   true,
		NextProtos:           []string{string(proto)},
		GetClientCertificate: certGetter,
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
