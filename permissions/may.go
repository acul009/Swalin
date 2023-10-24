package permissions

import (
	"context"
	"fmt"
	"rahnit-rmm/config"
	"rahnit-rmm/ent"
	"rahnit-rmm/ent/user"
	"rahnit-rmm/pki"
)

var ErrPermissionDenied = PermissionDeniedError{}

type PermissionDeniedError struct {
	PublicKey *pki.PublicKey
	Reason    string
}

func (e PermissionDeniedError) Error() string {
	return fmt.Sprintf("permission denied: %s", e.Reason)
}

func (e PermissionDeniedError) Is(target error) bool {
	_, ok := target.(PermissionDeniedError)
	return ok
}

func MayStartCommand(sender *pki.PublicKey, command string) error {
	isRoot, err := pki.IsRootPublicKey(sender)
	if err != nil {
		return fmt.Errorf("failed to check if public key is CA: %w", err)
	}

	if isRoot {
		return nil
	}

	db := config.DB()

	encoded, err := sender.Base64Encode()
	if err != nil {
		return fmt.Errorf("failed to encode public key: %w", err)
	}

	_, err = db.User.Query().Where(user.PublicKeyEQ(encoded)).Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return PermissionDeniedError{
				PublicKey: sender,
				Reason:    "requested sender is not a user",
			}
		}
		return fmt.Errorf("failed to query user: %w", err)
	}

	return nil

}
