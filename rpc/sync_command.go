package rpc

import (
	"fmt"
	"rahnit-rmm/util"
)

func NewSyncDownCommand[K comparable, T any](targetMap util.ObservableMap[K, T]) *syncDownCommand[K, T] {
	return &syncDownCommand[K, T]{
		targetMap: targetMap,
	}
}

type syncDownCommand[K comparable, T any] struct {
	targetMap util.ObservableMap[K, T]
}

type updateInfo[K comparable, T any] struct {
	delete bool
	Key    K
	Value  T
}

func (s *syncDownCommand[K, T]) ExecuteClient(session *RpcSession) error {

	initial := make(map[K]T)
	err := ReadMessage[map[K]T](session, initial)
	if err != nil {
		return fmt.Errorf("error reading message: %w", err)
	}

	for key, value := range initial {
		s.targetMap.Set(key, value)
	}

	update := &updateInfo[K, T]{}

	for {

		err := ReadMessage[*updateInfo[K, T]](session, update)
		if err != nil {
			return fmt.Errorf("error reading message: %w", err)
		}

		fmt.Printf("received update: %+v\n", update)

		if update.delete {
			s.targetMap.Delete(update.Key)
		} else {
			s.targetMap.Set(update.Key, update.Value)
		}
	}

}

func (s *syncDownCommand[K, T]) ExecuteServer(session *RpcSession) error {
	err := session.WriteResponseHeader(SessionResponseHeader{
		Code: 200,
		Msg:  "OK",
	})

	if err != nil {
		return fmt.Errorf("error writing response header: %w", err)
	}

	full := s.targetMap.GetAll()
	err = WriteMessage[map[K]T](session, full)
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}

	var updateErrChan = make(chan error)

	unsubscribe := s.targetMap.Subscribe(
		func(key K, value T) {
			err = WriteMessage[updateInfo[K, T]](session, updateInfo[K, T]{
				delete: false,
				Key:    key,
				Value:  value,
			})
			if err != nil {
				updateErrChan <- fmt.Errorf("error writing update message: %w", err)
			}
		},
		func(key K) {
			err = WriteMessage[updateInfo[K, T]](session, updateInfo[K, T]{
				delete: true,
				Key:    key,
			})
			if err != nil {
				updateErrChan <- fmt.Errorf("error writing update message: %w", err)
			}
		},
	)
	defer unsubscribe()

	err = <-updateErrChan

	return err
}
