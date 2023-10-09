package permissions

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"fmt"
	"rahnit-rmm/config"
	"rahnit-rmm/ent"
	"rahnit-rmm/ent/user"
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

func GetUserFromPublicKey(pub *ecdsa.PublicKey) (*ent.User, error) {
	db := config.DB()

	encoded, err := pki.EncodePubToString(pub)
	if err != nil {
		return nil, fmt.Errorf("failed to encode public key: %v", err)
	}

	userItem, err := db.User.Query().Where(user.PublicKeyEQ(encoded)).Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, PermissionDeniedError{
				PublicKey: pub,
				Reason:    "requested sender is not a user",
			}
		}
		return nil, fmt.Errorf("failed to query user: %v", err)
	}

	cert, err := pki.DecodeCertificate([]byte(userItem.Certificate))
	if err != nil {
		return nil, fmt.Errorf("failed to decode certificate: %v", err)
	}

	fmt.Println(cert)
	return nil, fmt.Errorf("not implemented")
}
