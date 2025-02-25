package tracker

import (
	"encoding/json"
	"sync"
)

type Storer[K comparable, V any] interface {
	Set(key K, value V)
	Get(key K) (V, bool)
	Delete(key K)
	Map(fn func(K, V))

	json.Marshaler
}

type store[K comparable, V any] struct {
	store map[K]V
	mutex sync.Mutex

	json.Marshaler
}

func NewStore[K comparable, V any]() Storer[K, V] {
	return &store[K, V]{
		store: make(map[K]V),
	}
}

func (s *store[K, V]) Set(key K, value V) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.store[key] = value
}

func (s *store[K, V]) Get(key K) (V, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	val, exists := s.store[key]
	return val, exists
}

func (s *store[K, V]) Delete(key K) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.store, key)
}

func (s *store[K, V]) Map(fn func(K, V)) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for k, v := range s.store {
		fn(k, v)
	}
}

func (s *store[K, V]) MarshalJSON() ([]byte, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return json.Marshal(s.store)
}
