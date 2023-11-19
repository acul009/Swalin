package util

import (
	"sync"

	"github.com/google/uuid"
)

type ObservableMap[K comparable, T any] interface {
	Subscribe(onSet func(K, T), onRemove func(K)) func()
	Set(key K, value T)
	Get(key K) (T, bool)
	Delete(key K)
	Size() int
	GetAll() map[K]T
	Update(key K, updateFunc func(value T, found bool) (T, bool))
}

type genericObservableMap[K comparable, T any] struct {
	m         map[K]T
	observers map[uuid.UUID]mapObserver[K, T]
	mutex     sync.RWMutex
}

type mapObserver[K comparable, T any] struct {
	set    func(K, T)
	delete func(K)
}

func NewObservableMap[K comparable, T any]() *genericObservableMap[K, T] {
	return &genericObservableMap[K, T]{
		m:         make(map[K]T),
		observers: make(map[uuid.UUID]mapObserver[K, T]),
		mutex:     sync.RWMutex{},
	}
}

func (m *genericObservableMap[K, T]) Subscribe(onSet func(K, T), onRemove func(K)) func() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	uuid := uuid.New()

	m.observers[uuid] = mapObserver[K, T]{
		set:    onSet,
		delete: onRemove,
	}

	return func() {
		m.mutex.Lock()
		defer m.mutex.Unlock()
		delete(m.observers, uuid)
	}
}

func (m *genericObservableMap[K, T]) Get(key K) (T, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	value, ok := m.m[key]
	return value, ok
}

func (m *genericObservableMap[K, T]) Has(key K) bool {
	_, ok := m.Get(key)
	return ok
}

func (m *genericObservableMap[K, T]) GetAll() map[K]T {
	copy := make(map[K]T)
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	for k, v := range m.m {
		copy[k] = v
	}
	return copy
}

func (m *genericObservableMap[K, T]) Set(key K, value T) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.m[key] = value
	for _, observer := range m.observers {
		observer.set(key, value)
	}
}

func (m *genericObservableMap[K, T]) Update(key K, updateFunc func(value T, found bool) (T, bool)) {

	m.mutex.Lock()
	defer m.mutex.Unlock()
	old, ok := m.m[key]
	new, changed := updateFunc(old, ok)
	if !changed {
		return
	}
	m.m[key] = new
	for _, observer := range m.observers {
		observer.set(key, m.m[key])
	}
}

func (m *genericObservableMap[K, T]) Delete(key K) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.m, key)
	for _, observer := range m.observers {
		observer.delete(key)
	}
}

func (m *genericObservableMap[K, T]) Size() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.m)
}

func (m *genericObservableMap[K, T]) ObserverCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.observers)
}
