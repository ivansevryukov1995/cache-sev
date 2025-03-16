package lru

import (
	"fmt"
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
	TTL      time.Duration
	Logger   pkg.Logger // Добавлено для логирования
}

func NewCache[KeyT comparable, ValueT any](capacity int, ttl time.Duration) *Cache[KeyT, ValueT] {
	return &Cache[KeyT, ValueT]{
		Capacity: capacity,
		Hash:     make(map[KeyT]*pkg.Node[KeyT, ValueT]),
		List:     pkg.NewList[KeyT, ValueT](),
		TTL:      ttl,
	}
}

// Get извлекает значение из кэша по заданному ключу.
// Возвращает значение и true, если ключ найден, иначе возвращает нулевое значение и false.
func (c *Cache[KeyT, ValueT]) Get(key KeyT) (ValueT, bool) {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	node, found := c.Hash[key]
	if found {
		// Проверяем, не истек ли срок действия элемента
		if time.Now().After(node.ExpiresAt) {
			// Элемент устарел, удаляем его
			c.removeAndDeleteNode(node)

			c.Logger.Log("Key expired: " + fmt.Sprintf("%v", key))

			var zeroValue ValueT
			return zeroValue, false
		}
		c.List.MoveToFront(node)

		c.Logger.Log("Retrieved key: " + fmt.Sprintf("%v", key))

		return node.Value, true
	}

	c.Logger.Log("Key not found: " + fmt.Sprintf("%v", key))

	var zeroValue ValueT
	return zeroValue, false
}

// Put добавляет новое значение в кэш по заданному ключу с установленным временем жизни.
// Если ключ уже существует, обновляет значение и перемещает его на переднюю позицию.
func (c *Cache[KeyT, ValueT]) Put(key KeyT, value ValueT) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	expiryTime := time.Now().Add(c.TTL) // Устанавливаем время истечения с учетом общего TTL

	if node, found := c.Hash[key]; found {
		// Обновляем значение и время истечения узла, перемещаем его на переднюю позицию
		node.Value = value
		node.ExpiresAt = expiryTime
		c.List.MoveToFront(node)

		c.Logger.Log("Updated key: " + fmt.Sprintf("%v", key))

		return
	}

	// Если кэш заполнен, нужно удалить последний элемент
	if len(c.Hash) >= c.Capacity {
		back := c.List.Back()
		if back != nil {
			c.removeAndDeleteNode(back)
			c.Logger.Log("Removed oldest key: " + fmt.Sprintf("%v", back.Key))
		}
	}

	// Создаем новый узел и добавляем его в кэш
	newNode := &pkg.Node[KeyT, ValueT]{
		Key:       key,
		Value:     value,
		ExpiresAt: expiryTime,
	}
	c.List.PushToFront(newNode)
	c.Hash[key] = newNode

	c.Logger.Log("Added key: " + fmt.Sprintf("%v", key))
}

// removeAndDeleteNode удаляет узел и его ключ из кэша.
func (c *Cache[KeyT, ValueT]) removeAndDeleteNode(node *pkg.Node[KeyT, ValueT]) {
	c.List.Remove(node)
	delete(c.Hash, node.Key)
}
