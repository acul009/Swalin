package rmm

import (
	"fmt"
	"rahnit-rmm/rpc"
	"rahnit-rmm/util"
)

func NewSyncDownCommand[K comparable, T any](targetMap util.ObservableMap[K, T]) *SyncDownCommand[K, T] {
	return &SyncDownCommand[K, T]{
		targetMap: targetMap,
	}
}

type SyncDownCommand[K comparable, T any] struct {
	targetMap util.ObservableMap[K, T]
}

type updateInfo[K comparable, T any] struct {
	Delete bool
	Key    K
	Value  T
}

func (s *SyncDownCommand[K, T]) ExecuteClient(session *rpc.RpcSession) error {

	initial := make(map[K]T)
	err := rpc.ReadMessage[map[K]T](session, initial)
	if err != nil {
		return fmt.Errorf("error reading message: %w", err)
	}

	for key, value := range initial {
		s.targetMap.Set(key, value)
	}

	update := &updateInfo[K, T]{}

	for {

		err := rpc.ReadMessage[*updateInfo[K, T]](session, update)
		if err != nil {
			return fmt.Errorf("error reading message: %w", err)
		}

		fmt.Printf("received update: %+v\n", update)

		if update.Delete {
			s.targetMap.Delete(update.Key)
		} else {
			s.targetMap.Set(update.Key, update.Value)
		}
	}

}

func (s *SyncDownCommand[K, T]) ExecuteServer(session *rpc.RpcSession) error {
	err := session.WriteResponseHeader(rpc.SessionResponseHeader{
		Code: 200,
		Msg:  "OK",
	})

	if err != nil {
		return fmt.Errorf("error writing response header: %w", err)
	}

	full := s.targetMap.GetAll()
	err = rpc.WriteMessage[map[K]T](session, full)
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}

	var updateErrChan = make(chan error)

	unsubscribe := s.targetMap.Subscribe(
		func(key K, value T) {
			err = rpc.WriteMessage[updateInfo[K, T]](session, updateInfo[K, T]{
				Delete: false,
				Key:    key,
				Value:  value,
			})
			if err != nil {
				updateErrChan <- fmt.Errorf("error writing update message: %w", err)
			}
		},
		func(key K, _ T) {
			err = rpc.WriteMessage[updateInfo[K, T]](session, updateInfo[K, T]{
				Delete: true,
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

func (s *SyncDownCommand[K, T]) SetMap(m util.ObservableMap[K, T]) {
	s.targetMap = m
}
