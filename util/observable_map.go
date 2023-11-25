package util

import (
	"sync"

	"github.com/google/uuid"
)

type Mappable[K comparable, T any] interface {
	Get(key K) (T, bool)
	Set(key K, value T)
	Delete(key K)
	Size() int
	GetAll() map[K]T
}

type mappableMap[K comparable, T any] struct {
	m map[K]T
}

func (m *mappableMap[K, T]) Get(key K) (T, bool) {
	value, ok := m.m[key]
	return value, ok
}

func (m *mappableMap[K, T]) Set(key K, value T) {
	m.m[key] = value
}

func (m *mappableMap[K, T]) Delete(key K) {
	delete(m.m, key)
}

func (m *mappableMap[K, T]) Size() int {
	return len(m.m)
}

func (m *mappableMap[K, T]) GetAll() map[K]T {
	return m.m
}

type ObservableMap[K comparable, T any] interface {
	Subscribe(onSet func(K, T), onRemove func(K, T)) func()
	Set(key K, value T)
	Get(key K) (T, bool)
	Delete(key K)
	Size() int
	GetAll() map[K]T
	Update(key K, updateFunc func(value T, found bool) (T, bool))
}

type genericObservableMap[K comparable, T any] struct {
	m             Mappable[K, T]
	observers     map[uuid.UUID]mapObserver[K, T]
	mutex         sync.RWMutex
	observerCount UpdateableObservable[int]
}

type mapObserver[K comparable, T any] struct {
	update func(K, T)
	delete func(K, T)
}

func NewObservableMap[K comparable, T any]() *genericObservableMap[K, T] {
	return &genericObservableMap[K, T]{
		m:             &mappableMap[K, T]{m: make(map[K]T)},
		observers:     make(map[uuid.UUID]mapObserver[K, T]),
		mutex:         sync.RWMutex{},
		observerCount: NewObservable[int](0),
	}
}

func (m *genericObservableMap[K, T]) Subscribe(onSet func(K, T), onRemove func(K, T)) func() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	uuid := uuid.New()

	m.observers[uuid] = mapObserver[K, T]{
		update: onSet,
		delete: onRemove,
	}

	m.updateObserverCount()

	return func() {
		m.mutex.Lock()
		defer m.mutex.Unlock()
		delete(m.observers, uuid)
		m.updateObserverCount()
	}
}

func (m *genericObservableMap[K, T]) Get(key K) (T, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.m.Get(key)
}

func (m *genericObservableMap[K, T]) Has(key K) bool {
	_, ok := m.Get(key)
	return ok
}

func (m *genericObservableMap[K, T]) GetAll() map[K]T {
	copy := make(map[K]T)
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	for k, v := range m.m.GetAll() {
		copy[k] = v
	}
	return copy
}

func (m *genericObservableMap[K, T]) Set(key K, value T) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.m.Set(key, value)
	for _, observer := range m.observers {
		observer.update(key, value)
	}
}

func (m *genericObservableMap[K, T]) Update(key K, updateFunc func(value T, found bool) (T, bool)) {

	m.mutex.Lock()
	defer m.mutex.Unlock()
	old, ok := m.m.Get(key)
	new, changed := updateFunc(old, ok)
	if !changed {
		return
	}
	m.m.Set(key, new)
	for _, observer := range m.observers {
		observer.update(key, new)
	}
}

func (m *genericObservableMap[K, T]) Delete(key K) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	value, ok := m.m.Get(key)
	if !ok {
		return
	}
	m.m.Delete(key)
	for _, observer := range m.observers {
		observer.delete(key, value)
	}
}

func (m *genericObservableMap[K, T]) Size() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.m.Size()
}

func (m *genericObservableMap[K, T]) ObserverCount() Observable[int] {
	return m.observerCount
}

func (m *genericObservableMap[K, T]) updateObserverCount() {
	old := m.observerCount.Get()
	new := len(m.observers)
	if old == new {
		return
	}
	m.observerCount.Update(func(i int) int {
		return new
	})
}
