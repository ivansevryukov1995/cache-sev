package lfu

import "github.com/ivansevryukov1995/cache-sev/pkg"

type Cache[KeyT comparable, ValueT any] struct {
	Capacity int
	Hash     map[int]*pkg.Set[KeyT]
	Values   map[KeyT]ValueT
	Freq     map[KeyT]int
	MinFreq  int
}

func NewCache[KeyT comparable, ValueT any](capacity int) *Cache[KeyT, ValueT] {
	return &Cache[KeyT, ValueT]{
		Capacity: capacity,
		Values:   make(map[KeyT]ValueT),
		Freq:     make(map[KeyT]int),
		Hash:     make(map[int]*pkg.Set[KeyT]),
		MinFreq:  0,
	}
}

func (c *Cache[KeyT, ValueT]) Get(key KeyT) (any, bool) {
	if _, ok := c.Values[key]; !ok {
		return nil, false
	}
	value := c.Values[key]
	freq := c.Freq[key]
	c.update(key, value, freq)
	return value, true
}

func (c *Cache[KeyT, ValueT]) Put(key KeyT, value ValueT) {
	if _, ok := c.Freq[key]; ok {
		freq := c.Freq[key]
		c.update(key, value, freq)
	} else {
		if len(c.Values) >= c.Capacity {
			c.evict()
		}
		c.Values[key] = value
		c.Freq[key] = 1
		if _, ok := c.Hash[1]; !ok {
			c.Hash[1] = pkg.NewSet[KeyT]()
		}
		c.Hash[1].Add(key)
		c.MinFreq = 1
	}
}

func (c *Cache[KeyT, ValueT]) update(key KeyT, value ValueT, freq int) {
	c.Hash[freq].Remove(key)
	if c.Hash[freq].IsEmpty() {
		delete(c.Hash, freq)
		if c.MinFreq == freq {
			c.MinFreq += 1
		}
	}

	c.Values[key] = value
	c.Freq[key] = freq + 1
	if _, ok := c.Hash[freq+1]; !ok {
		c.Hash[freq+1] = pkg.NewSet[KeyT]()
	}
	c.Hash[freq+1].Add(key)
}

func (c *Cache[KeyT, ValueT]) evict() {
	key := c.Hash[c.MinFreq].Pop()
	if c.Hash[c.MinFreq].IsEmpty() {
		delete(c.Hash, c.MinFreq)
	}
	delete(c.Values, key)
	delete(c.Freq, key)
}
