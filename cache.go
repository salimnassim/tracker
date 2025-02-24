package tracker

import (
	"encoding/json"
	"maps"
)

type Cacher[K comparable, V any] interface {
	Get(key K) (V, bool)
	FromStore(s Storer[K, V])
	Size() int

	json.Marshaler
}

type Cache[K comparable, V any] struct {
	cache map[K]V

	json.Marshaler
}

func NewCache[K comparable, V any]() Cacher[K, V] {
	return &Cache[K, V]{
		cache: make(map[K]V),
	}
}

func (r *Cache[K, V]) Get(key K) (V, bool) {
	val, exists := r.cache[key]
	return val, exists
}

func (r *Cache[K, V]) Size() int {
	return len(r.cache)
}

func (r *Cache[K, V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.cache)
}

func (c *Cache[K, V]) FromStore(s Storer[K, V]) {
	store, ok := s.(*Store[K, V])
	if !ok {
		c.cache = make(map[K]V)
		return
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	copyMap := make(map[K]V, len(store.store))
	maps.Copy(copyMap, store.store)
	c.cache = copyMap
}
