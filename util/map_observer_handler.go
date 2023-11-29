package util

import (
	"sync"

	"github.com/google/uuid"
)

type MapObserverHandler[K comparable, T any] struct {
	mutex         sync.RWMutex
	observers     map[uuid.UUID]mapObserver[K, T]
	observerCount UpdateableObservable[int]
}

type mapObserver[K comparable, T any] struct {
	update func(K, T)
	delete func(K, T)
}

func NewMapObserverHandler[K comparable, T any]() *MapObserverHandler[K, T] {
	return &MapObserverHandler[K, T]{
		observers:     make(map[uuid.UUID]mapObserver[K, T]),
		observerCount: newObservableWithoutObserverCount[int](0),
	}
}

func (m *MapObserverHandler[K, T]) Subscribe(onSet func(K, T), onRemove func(K, T)) func() {
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

func (m *MapObserverHandler[K, T]) ObserverCount() Observable[int] {
	return m.observerCount
}

func (m *MapObserverHandler[K, T]) updateObserverCount() {
	old := m.observerCount.Get()
	new := len(m.observers)
	if old == new {
		return
	}
	m.observerCount.Update(func(i int) int {
		return new
	})
}

func (m *MapObserverHandler[K, T]) NotifyUpdate(key K, value T) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	for _, observer := range m.observers {
		observer.update(key, value)
	}
}

func (m *MapObserverHandler[K, T]) NotifyDelete(key K, value T) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	for _, observer := range m.observers {
		observer.delete(key, value)
	}
}
