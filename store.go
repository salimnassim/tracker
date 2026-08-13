package tracker

import (
	"encoding/json"
	"maps"
	"sync"
	"sync/atomic"
)

type Storer[K comparable, V any] interface {
	Set(key K, value V)
	Get(key K) (V, bool)
	Delete(key K)
	Map(fn func(K, V))
	Snapshot() map[K]V

	json.Marshaler
}

type store[K comparable, V any] struct {
	data      map[K]V
	mutex     sync.RWMutex
	jsonCache atomic.Pointer[[]byte]
	dirty     atomic.Bool

	json.Marshaler
}

func NewStore[K comparable, V any]() Storer[K, V] {
	s := &store[K, V]{
		data: make(map[K]V),
	}
	s.dirty.Store(true)
	return s
}

func (s *store[K, V]) Set(key K, value V) {
	s.mutex.Lock()
	s.data[key] = value
	s.mutex.Unlock()
	s.dirty.Store(true)
}

func (s *store[K, V]) Get(key K) (V, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	val, exists := s.data[key]
	return val, exists
}

func (s *store[K, V]) Delete(key K) {
	s.mutex.Lock()
	delete(s.data, key)
	s.mutex.Unlock()
	s.dirty.Store(true)
}

func (s *store[K, V]) Map(fn func(K, V)) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for k, v := range s.data {
		fn(k, v)
	}
}

func (s *store[K, V]) Snapshot() map[K]V {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Create defensive copy
	snapshot := make(map[K]V, len(s.data))
	maps.Copy(snapshot, s.data)
	return snapshot
}

func (s *store[K, V]) MarshalJSON() ([]byte, error) {
	// Fast path: return cached JSON if available and clean
	if !s.dirty.Load() {
		if cached := s.jsonCache.Load(); cached != nil {
			return *cached, nil
		}
	}

	// Slow path: regenerate JSON
	s.mutex.RLock()
	jsonData, err := json.Marshal(s.data)
	s.mutex.RUnlock()

	if err == nil {
		s.jsonCache.Store(&jsonData)
		s.dirty.Store(false)
	}
	return jsonData, err
}
