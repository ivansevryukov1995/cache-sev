package lfu

import (
	"sync"

	"github.com/ivansevryukov1995/cache-sev/pkg"
)

// Cache структура описывает кэш с использованием алгоритма LFU
type Cache[KeyT comparable, ValueT any] struct {
	Capacity int
	Values   map[KeyT]ValueT
	Freq     map[KeyT]int
	Hash     map[int]*pkg.Set[KeyT]
	MinFreq  int
	Lock     sync.RWMutex
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

// Get извлекает значение по заданному ключу
func (c *Cache[KeyT, ValueT]) Get(key KeyT) (ValueT, bool) {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	value, exists := c.Values[key]
	if !exists {
		var zeroValue ValueT // Значение по умолчанию для типа ValueT
		return zeroValue, false
	}
	c.update(key, value)
	return value, true
}

// Put добавляет или обновляет значение для заданного ключа
func (c *Cache[KeyT, ValueT]) Put(key KeyT, value ValueT) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	if _, ok := c.Values[key]; ok {
		c.update(key, value)
		return
	}

	if len(c.Values) >= c.Capacity {
		c.evict()
	}

	c.Values[key] = value
	c.Freq[key] = 1
	c.addToFreqSet(key, 1)

	c.MinFreq = 1
}

// update обновляет частоту использования элемента
func (c *Cache[KeyT, ValueT]) update(key KeyT, value ValueT) {
	freq := c.Freq[key]
	c.Hash[freq].Remove(key)

	if c.Hash[freq].IsEmpty() {
		delete(c.Hash, freq)
		if c.MinFreq == freq {
			c.MinFreq++
		}
	}

	c.Values[key] = value
	c.Freq[key]++
	c.addToFreqSet(key, c.Freq[key])
}

// addToFreqSet добавляет ключ в соответствующий набор частоты
func (c *Cache[KeyT, ValueT]) addToFreqSet(key KeyT, freq int) {
	if _, exists := c.Hash[freq]; !exists {
		c.Hash[freq] = pkg.NewSet[KeyT]()
	}
	c.Hash[freq].Add(key)
}

// evict удаляет наименее используемый элемент
func (c *Cache[KeyT, ValueT]) evict() {
	if c.Hash[c.MinFreq].IsEmpty() {
		return // Ничего не делать, если набор пуст
	}

	key := c.Hash[c.MinFreq].Pop()
	delete(c.Values, key)
	delete(c.Freq, key)

	if c.Hash[c.MinFreq].IsEmpty() {
		delete(c.Hash, c.MinFreq)
	}
}
