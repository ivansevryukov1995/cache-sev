package lfu

import (
	"sync"
	"time"

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

// Get извлекает значение из кэша по заданному ключу.
// Возвращает значение и true, если ключ найден, иначе возвращает нулевое значение и false.
func (c *Cache[KeyT, ValueT]) Get(key KeyT) (ValueT, bool) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	if _, exists := c.Values[key]; !exists {
		var zeroValue ValueT // Значение по умолчанию для типа ValueT
		return zeroValue, false
	}

	value := c.Values[key]
	c.updateLocked(key, value)

	return value, true
}

// Put добавляет новое значение в кэш по заданному ключу с установленным временем жизни.
// Если ключ уже существует, обновляет значение, если ключ уже существует, если не существует
func (c *Cache[KeyT, ValueT]) Put(key KeyT, value ValueT, ttl time.Duration) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	// Обновляем значение, если ключ уже существует
	if _, ok := c.Freq[key]; ok {
		c.updateLocked(key, value)
		return
	}

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

	if ttl > 0 {
		go func() {
			<-time.After(ttl)
			c.Lock.Lock()
			defer c.Lock.Unlock()
			if _, exists := c.Values[key]; exists {
				c.removeLocked(key)

			}
		}()
	}
}

// updateLocked обновляет частоту использования элемента
func (c *Cache[KeyT, ValueT]) updateLocked(key KeyT, value ValueT) {
	freq := c.Freq[key]
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

// Наименее часто использовавшиеся (Least Frequently Used — LFU):
// убирает запись, которая использовалась наименее часто
func (c *Cache[KeyT, ValueT]) evict() {
	key := c.Hash[c.MinFreq].Pop()
	if c.Hash[c.MinFreq].IsEmpty() {
		delete(c.Hash, c.MinFreq)
	}
	delete(c.Values, key)
	delete(c.Freq, key)

}

// Remove удаляет элемент по ключу
func (c *Cache[KeyT, ValueT]) Remove(key KeyT) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	c.removeLocked(key)
}

// Метод removeLocked вызывается только под mutex
func (c *Cache[KeyT, ValueT]) removeLocked(key KeyT) {
	if freq, exists := c.Freq[key]; exists {
		c.Hash[freq].Remove(key)
		delete(c.Values, key)
		delete(c.Freq, key)

		if c.Hash[freq].IsEmpty() {
			delete(c.Hash, freq)
			if c.MinFreq == freq {
				c.MinFreq++
			}
		}

	}
}
