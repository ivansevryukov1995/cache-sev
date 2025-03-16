package lfu

import (
	"fmt"
	"sync"
	"time"
)

type Set[T comparable] struct {
	items map[T]struct{}
}

type Cache[KeyT comparable, ValueT any] struct {
	capacity      int
	values        map[KeyT]ValueT
	freq          map[KeyT]int
	keys          map[int]*Set[KeyT]
	minFreq       int
	ttl           time.Duration
	expiry        map[KeyT]time.Time
	mutex         sync.RWMutex
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		items: make(map[T]struct{}),
	}
}

func NewCache[KeyT comparable, ValueT any](capacity int, ttl time.Duration, cleanupInterval time.Duration) *Cache[KeyT, ValueT] {
	cache := &Cache[KeyT, ValueT]{
		capacity:      capacity,
		values:        make(map[KeyT]ValueT),
		freq:          make(map[KeyT]int),
		keys:          make(map[int]*Set[KeyT]),
		minFreq:       0,
		ttl:           ttl,
		expiry:        make(map[KeyT]time.Time),
		stopCleanup:   make(chan struct{}),
		cleanupTicker: time.NewTicker(cleanupInterval),
	}

	go cache.cleanupExpired()

	return cache
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
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if _, ok := cache.values[key]; !ok {
		return *new(ValueT), false
	}

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
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if _, ok := cache.freq[key]; ok {
		freq := cache.freq[key]
		cache.update(key, value, freq)
	} else {
		if len(cache.values) >= cache.capacity {
			cache.evict()
		}
		cache.values[key] = value
		cache.freq[key] = 1
		cache.expiry[key] = time.Now().Add(cache.ttl)
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
	cache.expiry[key] = time.Now().Add(cache.ttl)
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
		case <-cache.cleanupTicker.C:
			cache.mutex.Lock()
			for key, exp := range cache.expiry {
				if time.Now().After(exp) {
					cache.evictKey(key)
				}
			}
			cache.mutex.Unlock()
		case <-cache.stopCleanup:
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

func (cache *Cache[KeyT, ValueT]) StopCleanup() {
	close(cache.stopCleanup)
	cache.cleanupTicker.Stop()
}
