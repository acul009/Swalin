package util

import (
	"sync"

	"github.com/google/uuid"
)

type Observable[T any] interface {
	Set(T)
	Get() T
	Update(func(T) T)
	Subscribe(func(T)) func()
}

type GenericObservable[T any] struct {
	value     T
	observers map[uuid.UUID]func(T)
	mutex     sync.RWMutex
}

func NewGenericObservable[T any](value T) *GenericObservable[T] {
	return &GenericObservable[T]{
		value:     value,
		observers: make(map[uuid.UUID]func(T)),
		mutex:     sync.RWMutex{},
	}
}

func (o *GenericObservable[T]) Set(value T) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.value = value
	o.notifyObservers()
}

func (o *GenericObservable[T]) Get() T {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.value
}

func (o *GenericObservable[T]) Update(updateFunc func(T) T) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.value = updateFunc(o.value)
	o.notifyObservers()
}

func (o *GenericObservable[T]) notifyObservers() {
	for _, observer := range o.observers {
		observer(o.value)
	}
}

func (o *GenericObservable[T]) Subscribe(observer func(T)) func() {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	uuid := uuid.New()
	o.observers[uuid] = observer
	return func() {
		o.mutex.Lock()
		defer o.mutex.Unlock()
		delete(o.observers, uuid)
	}
}
