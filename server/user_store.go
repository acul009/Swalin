package server

import (
	"encoding/json"
	"fmt"

	"github.com/rahn-it/svalin/db"
	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/util"
)

type userStore struct {
	scope db.Scope
}

type user struct {
	certificate          *pki.Certificate
	encryptedPrivateKey  []byte
	clientHashingParams  *util.ArgonParameters
	serverHashingParams  *util.ArgonParameters
	doubleHashedPassword []byte
	totpSecret           []byte
}

func OpenUserStore(scope db.Scope) *userStore {
	return &userStore{
		scope: scope,
	}
}

const userPrefix = "user_"
const usernamePrefix = "username_"

func (us *userStore) SaveUser(u *user) error {
	publicKey := u.certificate.PublicKey().Base64Encode()
	username := u.certificate.GetName()

	raw, err := json.Marshal(u)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	err = us.scope.Update(func(b db.Bucket) error {
		err := b.Put([]byte(usernamePrefix+username), []byte(publicKey))
		if err != nil {
			return fmt.Errorf("failed to set username index: %w", err)
		}

		err = b.Put([]byte(userPrefix+publicKey), raw)
		if err != nil {
			return fmt.Errorf("failed to set user: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error during transaction: %w", err)
	}

	return nil
}

// GetUser retrieves a user with the given public key.
//
// It takes a publicKey of type *pki.PublicKey as a parameter.
// It returns a user pointer and an error.
//
// The function may return a nil user pointer without an error if no user is found.
func (u *userStore) GetUser(publicKey *pki.PublicKey) (*user, error) {
	encodedKey := publicKey.Base64Encode()
	var raw []byte
	err := u.scope.View(func(b db.Bucket) error {
		userData := b.Get([]byte(userPrefix + encodedKey))
		if userData == nil {
			return nil
		}

		raw := make([]byte, len(userData))
		copy(raw, userData)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error during transaction: %w", err)
	}

	user := &user{}

	err = json.Unmarshal(raw, user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return user, nil
}

func (u *userStore) GetUserByName(username string) (*user, error) {
	var raw []byte
	err := u.scope.View(func(b db.Bucket) error {
		userKey := b.Get([]byte(usernamePrefix + username))
		if userKey == nil {
			return fmt.Errorf("username not found")
		}

		userData := b.Get([]byte(userPrefix + string(userKey)))
		if userData == nil {
			return fmt.Errorf("user not found, index seems to be corrupted")
		}

		raw := make([]byte, len(userData))
		copy(raw, userData)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error during transaction: %w", err)
	}

	user := &user{}

	err = json.Unmarshal(raw, user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return user, nil
}

func (u *userStore) ForEach(fn func(*user) error) error {
	return u.scope.View(func(b db.Bucket) error {
		b.ForPrefix([]byte(userPrefix), func(k, v []byte) error {
			user := &user{}
			err := json.Unmarshal(v, user)
			if err != nil {
				return fmt.Errorf("failed to unmarshal user %s: %w", string(k), err)
			}

			return fn(user)
		})

		return nil
	})
}
