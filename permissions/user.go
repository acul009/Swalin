package permissions

import (
	"context"
	"fmt"
	"rahnit-rmm/config"
	"rahnit-rmm/ent"
	"rahnit-rmm/ent/user"
	"rahnit-rmm/pki"
)

func GetUserFromPublicKey(pub *pki.PublicKey) (*ent.User, error) {
	db := config.DB()

	encoded, err := pub.Base64Encode()
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

	return userItem, fmt.Errorf("not implemented")
}
