package util

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type UpdateableObservable[T any] interface {
	Observable[T]
	Update(func(T) T)
}

type Observable[T any] interface {
	Get() T
	Subscribe(func(T)) func()
}

type observable[T any] struct {
	value         T
	observers     map[uuid.UUID]func(T)
	mutex         sync.RWMutex
	observerCount UpdateableObservable[int]
}

func NewObservable[T any](value T) *observable[T] {
	return &observable[T]{
		value:         value,
		observers:     make(map[uuid.UUID]func(T)),
		mutex:         sync.RWMutex{},
		observerCount: newObservableWithoutObserverCount[int](0),
	}
}

func newObservableWithoutObserverCount[T any](value T) *observable[T] {
	return &observable[T]{
		value:         value,
		observers:     make(map[uuid.UUID]func(T)),
		mutex:         sync.RWMutex{},
		observerCount: nil,
	}
}

func (o *observable[T]) Get() T {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.value
}

func (o *observable[T]) Update(updateFunc func(T) T) {
	// log.Printf("updating observable...")
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.value = updateFunc(o.value)
	o.notifyObservers()
}

func (o *observable[T]) notifyObservers() {
	obs := o.observers
	for _, observer := range obs {
		observer(o.value)
	}
}

func (o *observable[T]) Subscribe(observer func(T)) func() {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	uuid := uuid.New()
	o.observers[uuid] = observer
	o.updateObserverCount()
	return func() {
		o.mutex.Lock()
		defer o.mutex.Unlock()
		delete(o.observers, uuid)
		o.updateObserverCount()
	}
}

func (o *observable[T]) updateObserverCount() {
	if o.observerCount == nil {
		return
	}

	old := o.observerCount.Get()
	new := len(o.observers)
	if old == new {
		return
	}
	o.observerCount.Update(func(_ int) int {
		return new
	})
}

func (o *observable[T]) ObserverCount() Observable[int] {
	return o.observerCount
}

type derivedObservable[T any, U any] struct {
	observable[U]
	upstream    Observable[T]
	transform   func(T) U
	unsubscribe func()
}

func DeriveObservable[T any, U any](upstream Observable[T], transform func(T) U) Observable[U] {
	if upstream == nil {
		panic(fmt.Errorf("observable is nil"))
	}

	if transform == nil {
		panic(fmt.Errorf("transform is nil"))
	}

	// log.Printf("deriving observable...")

	derived := &derivedObservable[T, U]{
		observable: observable[U]{
			observers: make(map[uuid.UUID]func(U)),
			mutex:     sync.RWMutex{},
		},
		upstream:  upstream,
		transform: transform,
	}

	return derived
}

func (o *derivedObservable[T, U]) Get() U {
	o.observable.mutex.RLock()
	defer o.observable.mutex.RUnlock()
	if o.unsubscribe != nil {
		return o.observable.Get()
	}
	return o.transform(o.upstream.Get())
}

func (o *derivedObservable[T, U]) Subscribe(observer func(U)) func() {
	o.observable.mutex.Lock()
	if len(o.observable.observers) == 0 {
		// log.Printf("connecting Observable to Upstream")
		o.unsubscribe = o.upstream.Subscribe(
			func(value T) {
				o.observable.value = o.transform(value)
				o.notifyObservers()
			},
		)
	}
	o.observable.mutex.Unlock()

	unsub := o.observable.Subscribe(observer)
	return func() {
		unsub()
		o.observable.mutex.Lock()
		if len(o.observable.observers) == 0 {
			o.unsubscribe()
			o.unsubscribe = nil
		}
		o.observable.mutex.Unlock()
	}
}
