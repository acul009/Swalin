package util

import (
	"sync"
)

type ObservableMap[K any, T any] interface {
	ForEach(f func(key K, value T) error) error
	Subscribe(onSet func(K, T), onRemove func(K, T)) func()
}

type UpdateableMap[K any, T any] interface {
	ObservableMap[K, T]
	Set(key K, value T)
	Get(key K) (T, bool)
	Delete(key K)
	Update(key K, updateFunc func(value T, found bool) (T, bool))
}

var _ UpdateableMap[any, any] = (*genericObservableMap[any, any])(nil)

type genericObservableMap[K comparable, T any] struct {
	m               map[K]T
	observerHandler *MapObserverHandler[K, T]
	mutex           sync.RWMutex
}

func NewObservableMap[K comparable, T any]() *genericObservableMap[K, T] {
	return &genericObservableMap[K, T]{
		observerHandler: NewMapObserverHandler[K, T](),
		m:               make(map[K]T),
		mutex:           sync.RWMutex{},
	}
}

func (m *genericObservableMap[K, T]) Subscribe(onSet func(K, T), onRemove func(K, T)) func() {
	return m.observerHandler.Subscribe(onSet, onRemove)
}

func (m *genericObservableMap[K, T]) ObserverCount() Observable[int] {
	return m.observerHandler.ObserverCount()
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

func (m *genericObservableMap[K, T]) ForEach(f func(key K, value T) error) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	for k, v := range m.m {
		err := f(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *genericObservableMap[K, T]) Set(key K, value T) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.m[key] = value
	m.observerHandler.NotifyDelete(key, value)
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
	m.observerHandler.NotifyUpdate(key, new)
}

func (m *genericObservableMap[K, T]) Delete(key K) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	value, ok := m.m[key]
	if !ok {
		return
	}
	delete(m.m, key)
	m.observerHandler.NotifyDelete(key, value)
}

func (m *genericObservableMap[K, T]) Size() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.m)
}
