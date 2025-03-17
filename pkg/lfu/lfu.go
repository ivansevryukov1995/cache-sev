package lfu

import (
	"fmt"
	"sync"
	"time"

	"github.com/ivansevryukov1995/cache-sev/pkg"
)

// CacheItem структура для хранения значения, частоты и времени истечения
type CacheItem[ValueT any] struct {
	Value     ValueT
	ExpiresAt time.Time
}

// Cache структура описывает кэш с использованием алгоритма LFU
type Cache[KeyT comparable, ValueT any] struct {
	Capacity int
	Values   map[KeyT]CacheItem[ValueT]
	Freq     map[KeyT]int
	Hash     map[int]*pkg.Set[KeyT]
	MinFreq  int
	Lock     sync.RWMutex
	TTL      time.Duration // Общий TTL для всех элементов
	Logger   pkg.Logger    // Добавлено для логирования
}

func NewCache[KeyT comparable, ValueT any](capacity int, ttl time.Duration) *Cache[KeyT, ValueT] {
	return &Cache[KeyT, ValueT]{
		Capacity: capacity,
		Values:   make(map[KeyT]CacheItem[ValueT]),
		Freq:     make(map[KeyT]int),
		Hash:     make(map[int]*pkg.Set[KeyT]),
		MinFreq:  0,
		TTL:      ttl,
	}
}

// Get извлекает значение по заданному ключу
func (c *Cache[KeyT, ValueT]) Get(key KeyT) (ValueT, bool) {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	item, exists := c.Values[key]
	if !exists || time.Now().After(item.ExpiresAt) {
		if exists {
			c.Logger.Log("Removing expired key: " + fmt.Sprintf("%v", key)) // Логирование удаления истекшего ключа
		}
		c.Remove(key)        // Удаляем элемент, если он истек
		var zeroValue ValueT // Значение по умолчанию для типа ValueT
		return zeroValue, false
	}
	c.update(key, item.Value)
	c.Logger.Log("Retrieved key: " + fmt.Sprintf("%v", key)) // Логирование успешного извлечения
	return item.Value, true
}

// Put добавляет или обновляет значение для заданного ключа
func (c *Cache[KeyT, ValueT]) Put(key KeyT, value ValueT) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	if _, ok := c.Values[key]; ok {
		c.update(key, value)                                   // Обновляем значение, если ключ уже существует
		c.Logger.Log("Updated key: " + fmt.Sprintf("%v", key)) // Логирование обновления ключа
		return
	}

	if len(c.Values) >= c.Capacity {
		c.evict()
		c.Logger.Log("Evicted an item from cache due to capacity") // Логирование удаления из-за переполнения
	}

	c.Values[key] = CacheItem[ValueT]{Value: value, ExpiresAt: time.Now().Add(c.TTL)}
	c.Freq[key] = 1
	c.addToFreqSet(key, 1)

	c.MinFreq = 1
	c.Logger.Log("Added key: " + fmt.Sprintf("%v", key)) // Логирование добавления ключа
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

	c.Values[key] = CacheItem[ValueT]{Value: value, ExpiresAt: time.Now().Add(c.TTL)} // Обновляем время истечения
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
	c.Logger.Log("Evicted key: " + fmt.Sprintf("%v", key)) // Логирование удаления наименее используемого ключа
}

// Remove удаляет элемент по ключу
func (c *Cache[KeyT, ValueT]) Remove(key KeyT) {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

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
		c.Logger.Log("Removed key: " + fmt.Sprintf("%v", key)) // Логирование удаления ключа
	}
}
