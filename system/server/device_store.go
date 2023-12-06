package server

import (
	"fmt"

	"github.com/rahn-it/svalin/db"
	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/util"
)

var _ util.ObservableMap[string, *pki.Certificate] = (*deviceStore)(nil)

type deviceStore struct {
	scope             db.Scope
	observableHandler *util.MapObserverHandler[string, *pki.Certificate]
}

func openDeviceStore(scope db.Scope) (*deviceStore, error) {
	return &deviceStore{
		scope:             scope,
		observableHandler: util.NewMapObserverHandler[string, *pki.Certificate](),
	}, nil
}

// Get retrieves a certificate from the device store based on the specified key.
//
// Parameters:
//   - key: the key used to identify the certificate in the store.
//
// Returns:
//   - *pki.Certificate: the retrieved certificate. If no certificate is found, it returns nil.
//   - error: any error that occurred during the retrieval process.
func (s *deviceStore) GetDevice(key *pki.PublicKey) (*pki.Certificate, error) {
	byteKey := []byte(key.Base64Encode())
	var raw []byte
	err := s.scope.View(func(b db.Bucket) error {
		found := b.Get(byteKey)
		if found != nil {
			raw = make([]byte, len(found))
			copy(raw, found)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error during transaction: %w", err)
	}

	if raw == nil {
		return nil, nil
	}

	cert, err := pki.CertificateFromPem(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal certificate: %w", err)
	}

	return cert, nil
}

func (s *deviceStore) ForEach(fn func(key string, value *pki.Certificate) error) error {
	return s.scope.View(func(b db.Bucket) error {
		return b.ForEach(func(k, v []byte) error {
			cert, err := pki.CertificateFromPem(v)
			if err != nil {
				return fmt.Errorf("failed to unmarshal certificate: %w", err)
			}

			return fn(string(k), cert)
		})
	})
}

func (s *deviceStore) Subscribe(onSet func(key string, value *pki.Certificate), onRemove func(key string, value *pki.Certificate)) func() {
	return s.observableHandler.Subscribe(onSet, onRemove)
}
