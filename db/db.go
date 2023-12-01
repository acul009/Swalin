package db

import (
	"fmt"
	"log"

	bolt "go.etcd.io/bbolt"
)

type DB struct {
	db      *bolt.DB
	context []byte
}

func Open(path string) (*DB, error) {

	b, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	var errChan <-chan error

	go func() {
		err = b.View(func(tx *bolt.Tx) error {
			errChan = tx.Check()
			return nil
		})
		if err != nil {
			panic(err)
		}
	}()

	errFound := false

	for checkErr := range errChan {
		if err != nil {
			log.Printf("db corrupted: %s", checkErr)
			errFound = true
		}
	}

	if errFound {
		return nil, fmt.Errorf("db corrupted")
	}

	return &DB{
		db: b,
	}, nil
}

func (d *DB) Scope(name []byte) Scope {
	return &scope{
		db:      d.db,
		context: d.context,
		path:    [][]byte{name},
	}
}
