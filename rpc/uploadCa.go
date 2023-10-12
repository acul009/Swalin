package rpc

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"rahnit-rmm/config"
	"rahnit-rmm/pki"
	"sync"
)

func UploadCaHandler() RpcCommand {
	return &UploadCa{}
}

type UploadCa struct {
	EncodedCa []byte
}

func UploadCaCmd() (*UploadCa, error) {
	ca, err := pki.GetRootCert()
	if err != nil {
		return nil, fmt.Errorf("failed to load CA certificate: %v", err)
	}

	encodedToPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: ca.Raw,
	})

	fmt.Println(string(encodedToPem))

	return &UploadCa{
		EncodedCa: encodedToPem,
	}, nil
}

var caMutex sync.Mutex = sync.Mutex{}

func (p *UploadCa) ExecuteClient(session *RpcSession) error {
	return nil
}

func (p *UploadCa) ExecuteServer(session *RpcSession) error {

	caMutex.Lock()
	defer caMutex.Unlock()

	_, err := pki.GetRootCert()
	if err == nil {
		session.WriteResponseHeader(SessionResponseHeader{
			Code: 409,
			Msg:  "CA certificate already exists",
		})
		return nil
	}
	if !errors.Is(err, pki.ErrNoRootCert) {
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

	// everything seems to be alright, we can save the CA certificate

	// create root user
	if session.Connection.role == RpcRoleServer {
		db, err := config.DB().Tx(context.Background())

		if err != nil {
			session.WriteResponseHeader(SessionResponseHeader{
				Code: 500,
				Msg:  "Failed to connect to database",
			})
			return fmt.Errorf("failed to connect to database: %v", err)
		}

		users, err := db.User.Query().All(context.Background())
		if err != nil {
			session.WriteResponseHeader(SessionResponseHeader{
				Code: 500,
				Msg:  "Failed to load users",
			})
			return fmt.Errorf("failed to load users: %v", err)
		}
		if len(users) > 0 {
			session.WriteResponseHeader(SessionResponseHeader{
				Code: 409,
				Msg:  "found already existing users",
			})
			return fmt.Errorf("found already existing users")
		}

		_, err = db.User.Create().SetCertificate(string(p.EncodedCa)).SetUsername("root").SetEncryptedPrivateKey("").SetPasswordDoubleHashed("").Save(context.Background())
		if err != nil {
			session.WriteResponseHeader(SessionResponseHeader{
				Code: 500,
				Msg:  "Failed to create root user",
			})
			return fmt.Errorf("failed to create root user: %v", err)
		}

		err = db.Commit()
		if err != nil {
			session.WriteResponseHeader(SessionResponseHeader{
				Code: 500,
				Msg:  "Failed to commit changes",
			})
			return fmt.Errorf("failed to commit changes: %v", err)
		}

	}

	// actually save the CA certificate
	err = pki.SaveRootCert(cert)
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
