package util

import "sync"

type syncedObservable[T any] struct {
	UpdateableObservable[T]
	register   func(UpdateableObservable[T])
	unregister func(UpdateableObservable[T])
	registered bool
	mutex      sync.Mutex
}

func NewSyncedObservable[T any](register func(UpdateableObservable[T]), unregister func(UpdateableObservable[T])) *syncedObservable[T] {

	var t T
	o := NewObservable[T](t)

	so := &syncedObservable[T]{
		UpdateableObservable: o,
		register:             register,
		unregister:           unregister,
		registered:           false,
		mutex:                sync.Mutex{},
	}

	o.ObserverCount().Subscribe(func(i int) {
		so.mutex.Lock()
		defer so.mutex.Unlock()
		if i > 0 {
			if !so.registered {
				so.register(so)
				so.registered = true
			}
		} else {
			if so.registered {
				so.unregister(so)
				so.registered = false
			}
		}
	})

	return so
}
