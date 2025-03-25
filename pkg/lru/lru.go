package lru

import (
	"sync"
	"time"

	"github.com/ivansevryukov1995/cache-sev/pkg"
)

// Структура Cache: Описывает, что кэш использует хеш-таблицу для быстрого доступа к элементам и
// двусвязный список для отслеживания порядка использования элементов.
type Cache[KeyT comparable, ValueT any] struct {
	Capacity int
	Hash     map[KeyT]*pkg.Node[KeyT, ValueT]
	List     *pkg.List[KeyT, ValueT]
	Lock     sync.RWMutex
}

func NewCache[KeyT comparable, ValueT any](capacity int) *Cache[KeyT, ValueT] {
	return &Cache[KeyT, ValueT]{
		Capacity: capacity,
		Hash:     make(map[KeyT]*pkg.Node[KeyT, ValueT]),
		List:     pkg.NewList[KeyT, ValueT](),
	}
}

// Get извлекает значение из кэша по заданному ключу.
// Возвращает значение и true, если ключ найден, иначе возвращает нулевое значение и false.
func (c *Cache[KeyT, ValueT]) Get(key KeyT) (ValueT, bool) {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	node, exists := c.Hash[key]
	if exists {
		c.List.MoveToFront(node)

		return node.Value, true
	}

	var zeroValue ValueT
	return zeroValue, false
}

// Put добавляет новое значение в кэш по заданному ключу с установленным временем жизни.
// Если ключ уже существует, обновляет значение и перемещает его на переднюю позицию.
func (c *Cache[KeyT, ValueT]) Put(key KeyT, value ValueT, ttl time.Duration) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	if node, ok := c.Hash[key]; ok {
		// Обновляем значение, перемещаем его на переднюю позицию
		node.Value = value
		c.List.MoveToFront(node)

		return
	}

	if len(c.Hash) >= c.Capacity {
		c.evict()
	}

	// Создаем новый узел и добавляем его в кэш
	newNode := &pkg.Node[KeyT, ValueT]{
		Key:   key,
		Value: value,
	}
	c.List.PushToFront(newNode)
	c.Hash[key] = newNode

	if ttl > 0 {
		go func() {
			<-time.After(ttl)
			c.Lock.Lock()
			defer c.Lock.Unlock()
			if _, exists := c.Hash[key]; exists {
				c.List.Remove(newNode)
				delete(c.Hash, key)

			}
		}()
	}
}

// Наиболее давно использовавшиеся (Least Recently Used – LRU):
// убирает запись, которая использовалась наиболее давно.
func (c *Cache[KeyT, ValueT]) evict() {
	back := c.List.Back()
	if back != nil {
		c.List.Remove(back)
		delete(c.Hash, back.Key)

	}
}
