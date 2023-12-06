package server

import (
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rahn-it/svalin/db"
	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/system"
	"github.com/rahn-it/svalin/util"
)

type userStore struct {
	scope db.Scope
}

func openUserStore(scope db.Scope) (*userStore, error) {
	return &userStore{
		scope: scope,
	}, nil
}

const userPrefix = "user_"
const usernamePrefix = "username_"

func (us *userStore) newUser(
	Certificate *pki.Certificate,
	EncryptedPrivateKey []byte,
	ClientHashingParams *util.ArgonParameters,
	ServerHashingParams *util.ArgonParameters,
	DoubleHashedPassword []byte,
	TotpSecret string,
) error {
	username := Certificate.GetName()
	publicKey := Certificate.PublicKey().Base64Encode()

	user := &system.User{
		Certificate:          Certificate,
		EncryptedPrivateKey:  EncryptedPrivateKey,
		ClientHashingParams:  ClientHashingParams,
		ServerHashingParams:  ServerHashingParams,
		DoubleHashedPassword: DoubleHashedPassword,
		TotpSecret:           TotpSecret,
	}

	raw, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	err = us.scope.Update(func(b db.Bucket) error {
		currentUserWithKey := b.Get([]byte(userPrefix + publicKey))
		if currentUserWithKey != nil {
			return errors.New("public key already in use")
		}

		currentUserWithName := b.Get([]byte(usernamePrefix + username))
		if currentUserWithName != nil {
			if Certificate.PublicKey().Base64Encode() != string(currentUserWithName) {
				return fmt.Errorf("username already in use")
			}
		}

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

// getUser retrieves a user with the given public key.
//
// It takes a publicKey of type *pki.PublicKey as a parameter.
// It returns a user pointer and an error.
//
// The function may return a nil user pointer without an error if no user is found.
func (u *userStore) getUser(publicKey *pki.PublicKey) (*system.User, error) {
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

	if raw == nil {
		return nil, nil
	}

	user := &system.User{}

	err = json.Unmarshal(raw, user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return user, nil
}

// getUserByName retrieves a user with the given username.
// if no user is found, nil is returned with no error.
func (u *userStore) getUserByName(username string) (*system.User, error) {
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

		raw = make([]byte, len(userData))
		copy(raw, userData)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error during transaction: %w", err)
	}

	user := &system.User{}

	err = json.Unmarshal(raw, user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return user, nil
}

func (u *userStore) forEach(fn func(*system.User) error) error {
	return u.scope.View(func(b db.Bucket) error {
		b.ForPrefix([]byte(userPrefix), func(k, v []byte) error {
			user := &system.User{}
			err := json.Unmarshal(v, user)
			if err != nil {
				return fmt.Errorf("failed to unmarshal user %s: %w", string(k), err)
			}

			return fn(user)
		})

		return nil
	})
}

var _ pki.Verifier = (*newUserVerifier)(nil)

type newUserVerifier struct {
	root          *pki.Certificate
	rootPool      *x509.CertPool
	intermediates *x509.CertPool
}

func newNewUserVerifier(root *pki.Certificate) (*newUserVerifier, error) {
	rootPool := x509.NewCertPool()
	rootPool.AddCert(root.ToX509())

	intermediates := x509.NewCertPool()

	return &newUserVerifier{
		root:          root,
		rootPool:      rootPool,
		intermediates: intermediates,
	}, nil
}

func (v *newUserVerifier) Verify(cert *pki.Certificate) ([]*pki.Certificate, error) {
	if cert.Equal(v.root) {
		return []*pki.Certificate{v.root}, nil
	}

	chain, err := cert.VerifyChain(v.rootPool, v.intermediates)
	if err != nil {
		return nil, fmt.Errorf("failed to verify certificate: %w", err)
	}

	certType := cert.Type()
	if certType != pki.CertTypeUser {
		return nil, fmt.Errorf("invalid certificate type: %s", certType)
	}

	return chain, nil
}

func (v *newUserVerifier) VerifyPublicKey(pub *pki.PublicKey) ([]*pki.Certificate, error) {
	return nil, errors.New("this verifier is not meant to be used for public keys")
}
