package system

import (
	"github.com/rahn-it/svalin/db"
	"github.com/rahn-it/svalin/pki"
	"go.etcd.io/bbolt"
)

func LoadHostCredentials(scope db.Scope) (*pki.PermanentCredentials, error) {
	scope.View(func(b *bbolt.Bucket) error {

	})
}
