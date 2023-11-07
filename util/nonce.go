package util

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

type Nonce []byte

func NewNonce() (Nonce, error) {
	nonce := make([]byte, 32)
	n, err := rand.Reader.Read(nonce)
	if err != nil {
		return nil, fmt.Errorf("error generating nonce: %w", err)
	}
	if n != len(nonce) {
		return nil, fmt.Errorf("error generating nonce: %w", err)
	}
	return Nonce(nonce), nil
}

type NonceStorage struct {
	nonceMap map[string]int64
	mutex    sync.RWMutex
}

func NewNonceStorage() *NonceStorage {
	return &NonceStorage{
		nonceMap: make(map[string]int64),
		mutex:    sync.RWMutex{},
	}
}

func (s *NonceStorage) CheckNonce(nonce Nonce) bool {
	key := base64.StdEncoding.EncodeToString(nonce)
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	_, ok := s.nonceMap[key]
	return !ok
}

func (s *NonceStorage) AddNonce(nonce Nonce) {
	key := base64.StdEncoding.EncodeToString(nonce)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.nonceMap[key] = time.Now().Unix()
}

func (s *NonceStorage) Cleanup(expiration int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for key, timestamp := range s.nonceMap {
		if timestamp < time.Now().Unix()-expiration {
			delete(s.nonceMap, key)
		}
	}
}
