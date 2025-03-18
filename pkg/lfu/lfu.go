package lfu

import (
	"fmt"
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
	Logger   pkg.Logger // Добавлено для логирования
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
	c.Lock.Lock()
	defer c.Lock.Unlock()

	if _, exists := c.Values[key]; !exists {
		var zeroValue ValueT // Значение по умолчанию для типа ValueT
		return zeroValue, false
	}

	value := c.Values[key]
	freq := c.Freq[key]
	c.updateLocked(key, value, freq)

	c.Logger.Log("Retrieved key: " + fmt.Sprintf("%v", key)) // Логирование успешного извлечения

	return value, true
}

func (c *Cache[KeyT, ValueT]) Put(key KeyT, value ValueT, ttl time.Duration) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	if _, ok := c.Freq[key]; ok {
		freq := c.Freq[key]
		c.updateLocked(key, value, freq)                      // Обновляем значение, если ключ уже существует
		c.Logger.Log("Update key: " + fmt.Sprintf("%v", key)) // Логирование обновления ключа
	} else {
		// Здесь испоняется политика вытеснения ключа из кеша LFU
		// Если кэш заполнен, нужно удалить наименее используемый элемент.
		if len(c.Values) >= c.Capacity {
			c.evict()
			c.Logger.Log("Evicted an item from cache due to capacity") // Логирование удаления из-за переполнения
		}
		c.Values[key] = value
		c.Freq[key] = 1
		if _, ok := c.Hash[1]; !ok {
			c.Hash[1] = pkg.NewSet[KeyT]()
		}
		c.Hash[1].Add(key)
		c.MinFreq = 1
		c.Logger.Log("Added key: " + fmt.Sprintf("%v", key)) // Логирование добавления ключа

		if ttl > 0 {
			go func() {
				<-time.After(ttl)
				c.Lock.Lock()
				defer c.Lock.Unlock()
				if _, exists := c.Values[key]; exists {
					c.removeLocked(key)
					c.Logger.Log("Removed key with ttl expired: " + fmt.Sprintf("%v", key))
				}
			}()
		}
	}
}

// updateLocked обновляет частоту использования элемента
func (c *Cache[KeyT, ValueT]) updateLocked(key KeyT, value ValueT, freq int) {
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

	c.Logger.Log(fmt.Sprintf("Key: %v, Frequency: %v", key, freq+1)) // Логирование частоты использования ключа
}

// evict удаляет наименее используемый элемент
func (c *Cache[KeyT, ValueT]) evict() {
	key := c.Hash[c.MinFreq].Pop()
	if c.Hash[c.MinFreq].IsEmpty() {
		delete(c.Hash, c.MinFreq)
	}
	delete(c.Values, key)
	delete(c.Freq, key)
	c.Logger.Log("Evicted key: " + fmt.Sprintf("%v", key)) // Логирование удаления наименее используемого ключа
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
		c.Logger.Log("Removed key: " + fmt.Sprintf("%v", key)) // Логирование удаления ключа
	}
}
