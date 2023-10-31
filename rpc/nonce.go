package rpc

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

type nonceStorage struct {
	nonceMap map[string]int64
	mutex    sync.RWMutex
}

func NewNonceStorage() *nonceStorage {
	return &nonceStorage{
		nonceMap: make(map[string]int64),
		mutex:    sync.RWMutex{},
	}
}

func (s *nonceStorage) CheckNonce(nonce Nonce) bool {
	key := base64.StdEncoding.EncodeToString(nonce)
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	_, ok := s.nonceMap[key]
	return !ok
}

func (s *nonceStorage) AddNonce(nonce Nonce) {
	key := base64.StdEncoding.EncodeToString(nonce)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.nonceMap[key] = time.Now().Unix()
}

func (s *nonceStorage) cleanup() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for key, timestamp := range s.nonceMap {
		if timestamp < time.Now().Unix()-messageExpiration*2 {
			delete(s.nonceMap, key)
		}
	}
}
