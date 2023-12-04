package db

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type Scope interface {
	Update(fn func(b Bucket) error) error
	View(fn func(b Bucket) error) error
	Scope(name string) Scope
}

var _ Scope = (*scope)(nil)

type scope struct {
	db      *bolt.DB
	context []byte
	path    [][]byte
}

func getBucket(tx *bolt.Tx, context []byte, path [][]byte) (*bucket, error) {
	bucket := tx.Bucket(context)
	if bucket == nil {
		return nil, fmt.Errorf("failed to get context bucket")
	}

	for _, p := range path {
		bucket = bucket.Bucket(p)
		if bucket == nil {
			return nil, fmt.Errorf("failed to get path bucket")
		}
	}

	return newBucket(bucket), nil
}

func (s *scope) Update(fn func(s Bucket) error) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := getBucket(tx, s.context, s.path)
		if err != nil {
			return fmt.Errorf("failed to access scope: %w", err)
		}
		return fn(bucket)
	})
}

func (s *scope) View(fn func(s Bucket) error) error {
	return s.db.View(func(tx *bolt.Tx) error {
		bucket, err := getBucket(tx, s.context, s.path)
		if err != nil {
			return fmt.Errorf("failed to access scope: %w", err)
		}
		return fn(bucket)
	})
}

func (s *scope) Scope(name string) Scope {
	key := []byte(name)
	subPath := make([][]byte, len(s.path), len(s.path)+1)
	copy(subPath, s.path)
	subPath = append(subPath, key)

	err := s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := getBucket(tx, s.context, s.path)
		if err != nil {
			return fmt.Errorf("failed to access scope: %w", err)
		}

		_, err = bucket.CreateBucketIfNotExists(key)
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
