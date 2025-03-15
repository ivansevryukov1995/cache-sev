package lfu

import "fmt"

type Set[T comparable] struct {
	items map[T]struct{}
}

type Cache[KeyT comparable, ValueT any] struct {
	capacity int
	values   map[KeyT]ValueT
	freq     map[KeyT]int
	keys     map[int]*Set[KeyT]
	minFreq  int
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		items: make(map[T]struct{}),
	}
}

func NewCache[KeyT comparable, ValueT any](capacity int) *Cache[KeyT, ValueT] {
	return &Cache[KeyT, ValueT]{
		capacity: capacity,
		values:   make(map[KeyT]ValueT),
		freq:     make(map[KeyT]int),
		keys:     make(map[int]*Set[KeyT]),
		minFreq:  0,
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
		panic("Pop() from the emty Set")
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

func (cache *Cache[KeyT, ValueT]) Get(key KeyT) (any, bool) {
	if _, ok := cache.values[key]; !ok {
		return nil, false
	}
	value := cache.values[key]
	freq := cache.freq[key]
	cache.update(key, value, freq)
	return value, true
}

func (cache *Cache[KeyT, ValueT]) Put(key KeyT, value ValueT) {
	if _, ok := cache.freq[key]; ok {
		freq := cache.freq[key]
		cache.update(key, value, freq)
	} else {
		if len(cache.values) >= cache.capacity {
			cache.evict()
		}
		cache.values[key] = value
		cache.freq[key] = 1
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
}
