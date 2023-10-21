package permissions

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"rahnit-rmm/config"
	"rahnit-rmm/ent"
	"rahnit-rmm/ent/user"
	"rahnit-rmm/pki"
)

func GetUserFromPublicKey(pub *ecdsa.PublicKey) (*ent.User, error) {
	db := config.DB()

	encoded, err := pki.EncodePubToString(pub)
	if err != nil {
		return nil, fmt.Errorf("failed to encode public key: %w", err)
	}

	userItem, err := db.User.Query().Where(user.PublicKeyEQ(encoded)).Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, PermissionDeniedError{
				PublicKey: pub,
				Reason:    "requested sender is not a user",
			}
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	cert, err := pki.DecodeCertificate([]byte(userItem.Certificate))
	if err != nil {
		return nil, fmt.Errorf("failed to decode certificate: %w", err)
	}

	fmt.Println(cert)
	return nil, fmt.Errorf("not implemented")
}
