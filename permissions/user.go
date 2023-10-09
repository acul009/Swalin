package permissions

import (
	"context"
	"crypto/x509"
	"fmt"
	"rahnit-rmm/config"
	"rahnit-rmm/ent"
	"rahnit-rmm/pki"
)

func CreateUserEntry(username string, cert *x509.Certificate) (*ent.User, error) {
	db := config.DB()

	pubEncoded, err := pki.GetEncodedPublicKey(cert)
	if err != nil {
		return nil, err
	}

	certEncoded := pki.EncodeCertificate(cert)

	user, err := db.User.Create().SetPublicKey(pubEncoded).SetUsername(username).SetCertificate(string(certEncoded)).Save(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return user, nil
}
