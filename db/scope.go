package db

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type Scope interface {
	Update(fn func(b *bolt.Bucket) error) error
	View(fn func(b *bolt.Bucket) error) error
	Scope(name []byte) Scope
}

var _ Scope = (*scope)(nil)

type scope struct {
	db      *bolt.DB
	context []byte
	path    [][]byte
}

func getBucket(tx *bolt.Tx, context []byte, path [][]byte) (*bolt.Bucket, error) {
	bucket, err := tx.CreateBucketIfNotExists(context)
	if err != nil {
		return nil, fmt.Errorf("failed to create context bucket: %w", err)
	}
	for _, p := range path {
		bucket, err = bucket.CreateBucketIfNotExists(p)
		if err != nil {
			return nil, fmt.Errorf("failed to create path bucket: %w", err)
		}
	}
	return bucket, nil
}

func (s *scope) Update(fn func(s *bolt.Bucket) error) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := getBucket(tx, s.context, s.path)
		if err != nil {
			return fmt.Errorf("failed to access scope: %w", err)
		}
		return fn(bucket)
	})
}

func (s *scope) View(fn func(s *bolt.Bucket) error) error {
	return s.db.View(func(tx *bolt.Tx) error {
		bucket, err := getBucket(tx, s.context, s.path)
		if err != nil {
			return fmt.Errorf("failed to access scope: %w", err)
		}
		return fn(bucket)
	})
}

func (s *scope) Scope(name []byte) Scope {
	subPath := make([][]byte, len(s.path), len(s.path)+1)
	copy(subPath, s.path)
	subPath = append(subPath, name)

	err := s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := getBucket(tx, s.context, s.path)
		if err != nil {
			return fmt.Errorf("failed to access scope: %w", err)
		}

		_, err = bucket.CreateBucketIfNotExists(name)
		if err != nil {
			return fmt.Errorf("failed to create scope: %w", err)
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	return &scope{
		db:      s.db,
		context: s.context,
		path:    subPath,
	}
}
