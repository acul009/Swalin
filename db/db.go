package db

import (
	"fmt"

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

	// TODO: enable checks

	// var errChan <-chan error

	// go func() {
	// 	err = b.View(func(tx *bolt.Tx) error {
	// 		errChan = tx.Check()
	// 		return nil
	// 	})
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }()

	// errFound := false

	// for checkErr := range errChan {
	// 	if err != nil {
	// 		log.Printf("db corrupted: %s", checkErr)
	// 		errFound = true
	// 	}
	// }

	// if errFound {
	// 	return nil, fmt.Errorf("db corrupted")
	// }

	return &DB{
		db: b,
	}, nil
}

func (d *DB) ContextList() ([][]byte, error) {
	list := make([][]byte, 0)
	err := d.db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			list = append(list, name)
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list context: %w", err)
	}

	return list, nil
}

func (d *DB) Context(context []byte) Scope {
	err := d.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(context)
		if err != nil {
			return fmt.Errorf("failed to create context: %w", err)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	return &scope{
		db:      d.db,
		context: context,
		path:    [][]byte{},
	}
}
