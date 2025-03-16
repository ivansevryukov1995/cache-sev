package lfu

import (
	"fmt"
	"time"

	"github.com/ivansevryukov1995/cache-sev/pkg/cache"
)

type Set[T comparable] struct {
	items map[T]struct{}
}

type Cache[KeyT comparable, ValueT any] struct {
	base     *cache.BaseCache[KeyT, ValueT] // Использование базовой структуры
	capacity int
	values   map[KeyT]ValueT
	freq     map[KeyT]int
	keys     map[int]*Set[KeyT]
	minFreq  int
	expiry   map[KeyT]time.Time
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{items: make(map[T]struct{})}
}

func NewCache[KeyT comparable, ValueT any](capacity int, ttl time.Duration, cleanupInterval time.Duration) *Cache[KeyT, ValueT] {
	base := cache.NewBaseCache[KeyT, ValueT](ttl, cleanupInterval)
	return &Cache[KeyT, ValueT]{
		base:     base,
		capacity: capacity,
		values:   make(map[KeyT]ValueT),
		freq:     make(map[KeyT]int),
		keys:     make(map[int]*Set[KeyT]),
		expiry:   make(map[KeyT]time.Time),
	}
}

func (s *Set[T]) Add(val T) {
	s.items[val] = struct{}{}
}

func (s *Set[T]) Remove(val T) {
	delete(s.items, val)
}

func (s *Set[T]) Pop() T {
	if len(s.items) == 0 {
		panic("Pop() from the empty Set")
	}
	for key := range s.items {
		s.Remove(key)
		return key
	}
	var ans T
	return ans
}

func (s *Set[T]) IsEmpty() bool {
	return len(s.items) == 0
}

func (s *Set[T]) String() string {
	return fmt.Sprintf("%v", s.items)
}

func (cache *Cache[KeyT, ValueT]) Get(key KeyT) (ValueT, bool) {
	cache.base.Mutex.Lock()
	defer cache.base.Mutex.Unlock()

	if _, ok := cache.values[key]; !ok {
		return *new(ValueT), false
	}

	// Проверяем, истек ли ключ
	if time.Now().After(cache.expiry[key]) {
		cache.evictKey(key)
		return *new(ValueT), false
	}

	value := cache.values[key]
	freq := cache.freq[key]
	cache.update(key, value, freq)
	return value, true
}

func (cache *Cache[KeyT, ValueT]) Put(key KeyT, value ValueT) {
	cache.base.Mutex.Lock()
	defer cache.base.Mutex.Unlock()

	// Если ключ уже есть, обновляем его
	if freq, found := cache.freq[key]; found {
		cache.update(key, value, freq)
	} else {
		// Ключа нет, нужно добавить его
		if len(cache.values) >= cache.capacity {
			cache.evict()
		}
		cache.values[key] = value
		cache.freq[key] = 1
		cache.expiry[key] = time.Now().Add(cache.base.Ttl)
		if _, ok := cache.keys[1]; !ok {
			cache.keys[1] = NewSet[KeyT]()
		}
		cache.keys[1].Add(key)
		cache.minFreq = 1
	}
}

func (cache *Cache[KeyT, ValueT]) update(key KeyT, value ValueT, freq int) {
	cache.keys[freq].Remove(key)
	if cache.keys[freq].IsEmpty() {
		delete(cache.keys, freq)
		if cache.minFreq == freq {
			cache.minFreq += 1
		}
	}

	cache.values[key] = value
	cache.expiry[key] = time.Now().Add(cache.base.Ttl)
	cache.freq[key] = freq + 1
	if _, ok := cache.keys[freq+1]; !ok {
		cache.keys[freq+1] = NewSet[KeyT]()
	}
	cache.keys[freq+1].Add(key)
}

func (cache *Cache[KeyT, ValueT]) evict() {
	key := cache.keys[cache.minFreq].Pop()
	if cache.keys[cache.minFreq].IsEmpty() {
		delete(cache.keys, cache.minFreq)
	}
	delete(cache.values, key)
	delete(cache.freq, key)
	delete(cache.expiry, key)
}

func (cache *Cache[KeyT, ValueT]) cleanupExpired() {
	for {
		select {
		case <-cache.base.CleanupTicker.C:
			cache.base.Mutex.Lock()
			for key, exp := range cache.expiry {
				if time.Now().After(exp) {
					cache.evictKey(key)
				}
			}
			cache.base.Mutex.Unlock()
		case <-cache.base.StopCleanUp:
			return
		}
	}
}

func (cache *Cache[KeyT, ValueT]) evictKey(key KeyT) {
	if _, ok := cache.values[key]; ok {
		freq := cache.freq[key]
		cache.keys[freq].Remove(key)

		if cache.keys[freq].IsEmpty() {
			delete(cache.keys, freq)
			if cache.minFreq == freq {
				cache.minFreq += 1
			}
		}

		delete(cache.values, key)
		delete(cache.freq, key)
		delete(cache.expiry, key)
	}
}
