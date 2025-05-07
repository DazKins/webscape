package util

import (
	"fmt"
	"sync"
)

// BiMap is a bi-directional map that allows efficient lookups in both directions.
// K is the type for keys, V is the type for values.
type BiMap[K comparable, V comparable] struct {
	forward  map[K]V
	backward map[V]K
	mu       sync.RWMutex
}

// NewBiMap creates a new empty BiMap.
func NewBiMap[K comparable, V comparable]() *BiMap[K, V] {
	return &BiMap[K, V]{
		forward:  make(map[K]V),
		backward: make(map[V]K),
	}
}

func (b *BiMap[K, V]) Put(key K, value V) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if oldValue, exists := b.forward[key]; exists {
		delete(b.backward, oldValue)
	}
	if oldKey, exists := b.backward[value]; exists {
		delete(b.forward, oldKey)
	}

	b.forward[key] = value
	b.backward[value] = key
}

func (b *BiMap[K, V]) Get(key K) (V, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	value, exists := b.forward[key]
	return value, exists
}

func (b *BiMap[K, V]) GetKey(value V) (K, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	key, exists := b.backward[value]
	return key, exists
}

func (b *BiMap[K, V]) Delete(key K) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if value, exists := b.forward[key]; exists {
		delete(b.forward, key)
		delete(b.backward, value)
	}
}

func (b *BiMap[K, V]) DeleteValue(value V) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if key, exists := b.backward[value]; exists {
		delete(b.backward, value)
		delete(b.forward, key)
	}
}

func (b *BiMap[K, V]) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.forward)
}

func (b *BiMap[K, V]) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.forward = make(map[K]V)
	b.backward = make(map[V]K)
}

func (b *BiMap[K, V]) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return fmt.Sprintf("BiMap{%v}", b.forward)
}
