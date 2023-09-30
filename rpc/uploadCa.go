package rpc

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"rahnit-rmm/pki"
)

func UploadCaHandler() RpcCommand {
	return &UploadCa{}
}

type UploadCa struct {
	EncodedCa []byte
}

func UploadCaCmd() (*UploadCa, error) {
	ca, err := pki.GetCaCert()
	if err != nil {
		return nil, fmt.Errorf("failed to load CA certificate: %v", err)
	}

	fmt.Printf("CA certificate:\n%v\n", ca.Raw)

	encodedToPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: ca.Raw,
	})

	fmt.Println(string(encodedToPem))

	return &UploadCa{
		EncodedCa: encodedToPem,
	}, nil
}

func (p *UploadCa) ExecuteClient(session *RpcSession) error {
	return nil
}

func (p *UploadCa) ExecuteServer(session *RpcSession) error {
	_, err := pki.GetCaCert()
	if err == nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 409,
			Msg:  "CA certificate already exists",
		})
		return nil
	}
	if !errors.Is(err, pki.ErrNoCaCert) {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 500,
			Msg:  "Failed to load CA certificate",
		})
		return fmt.Errorf("failed to load CA certificate: %v", err)
	}

	// Decode the PEM-encoded certificate
	block, _ := pem.Decode(p.EncodedCa)
	if block == nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 422,
			Msg:  "failed to decode certificate PEM",
		})
		return fmt.Errorf("failed to decode certificate PEM")
	}

	// Parse the CA certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 422,
			Msg:  "Failed to parse CA certificate",
		})
		return fmt.Errorf("failed to parse CA certificate: %v", err)
	}

	err = pki.SaveCaCert(cert)
	if err != nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 500,
			Msg:  "Failed to save CA certificate",
		})
		return fmt.Errorf("failed to save CA certificate: %v", err)
	}

	session.WriteResponseHeader(SessionResponseHeader{
		Code: 200,
		Msg:  "OK",
	})
	return nil
}

func (p *UploadCa) GetKey() string {
	return "uploadCa"
}
