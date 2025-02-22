package tracker

import (
	"encoding/json"
	"sync"
)

type Storer[K comparable, V any] interface {
	Set(key K, value V)
	Get(key K) (V, bool)
	Delete(key K)
	json.Marshaler
}

type Store[K comparable, V any] struct {
	store map[K]V
	mutex sync.Mutex
}

func NewStore[K comparable, V any]() Storer[K, V] {
	return &Store[K, V]{
		store: make(map[K]V),
	}
}

func (s *Store[K, V]) Set(key K, value V) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.store[key] = value
}

func (s *Store[K, V]) Get(key K) (V, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	val, exists := s.store[key]
	return val, exists
}

func (s *Store[K, V]) Delete(key K) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.store, key)
}

func (s *Store[K, V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.store)
}
