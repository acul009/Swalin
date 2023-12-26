package system

import (
	"fmt"

	"github.com/rahn-it/svalin/rpc"
	"github.com/rahn-it/svalin/util"
)

func NewSyncDownCommand[K comparable, T any](targetMap util.UpdateableMap[K, T]) *SyncDownCommand[K, T] {
	return &SyncDownCommand[K, T]{
		targetMap: targetMap,
	}
}

type SyncDownCommand[K comparable, T any] struct {
	targetMap util.UpdateableMap[K, T]
	sourceMap util.ObservableMap[K, T]
}

type updateInfo[K comparable, T any] struct {
	Delete bool
	Key    K
	Value  T
}

func (s *SyncDownCommand[K, T]) ExecuteClient(session *rpc.RpcSession) error {
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

	err = s.sourceMap.ForEach(func(key K, value T) error {
		return rpc.WriteMessage[updateInfo[K, T]](session, updateInfo[K, T]{
			Delete: false,
			Key:    key,
			Value:  value,
		})
	})

	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}

	var updateErrChan = make(chan error)

	unsubscribe := s.sourceMap.Subscribe(
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

func (s *SyncDownCommand[K, T]) SetSourceMap(m util.ObservableMap[K, T]) {
	s.sourceMap = m
}
