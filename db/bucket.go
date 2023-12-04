package db

import (
	"bytes"

	bolt "go.etcd.io/bbolt"
)

type Bucket interface {
	Get(key []byte) []byte
	Put(key, value []byte) error
	ForEach(func(k, v []byte) error) error
	ForPrefix(prefix []byte, fn func(k, v []byte) error) error
}

type bucket struct {
	*bolt.Bucket
}

func newBucket(b *bolt.Bucket) *bucket {
	return &bucket{
		Bucket: b,
	}
}

func (b *bucket) ForPrefix(prefix []byte, fn func(k, v []byte) error) error {
	cursor := b.Cursor()
	for k, v := cursor.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = cursor.Next() {
		err := fn(k, v)
		if err != nil {
			return err
		}
	}

	return nil
}
