package lru

import (
	"sync"
	"time"

	"github.com/ivansevryukov1995/cache-sev/pkg"
)

type DataNode[KeyT comparable, ValueT any] struct {
	Key   KeyT
	Value ValueT
	Prev  *DataNode[KeyT, ValueT]
	Next  *DataNode[KeyT, ValueT]
}

func (n *DataNode[KeyT, ValueT]) GetPrev() pkg.NodeInterface[KeyT, ValueT] {
	return n.Prev
}

func (n *DataNode[KeyT, ValueT]) GetNext() pkg.NodeInterface[KeyT, ValueT] {
	return n.Next
}

func (n *DataNode[KeyT, ValueT]) SetPrev(prev pkg.NodeInterface[KeyT, ValueT]) {
	n.Prev = prev.(*DataNode[KeyT, ValueT])
}

func (n *DataNode[KeyT, ValueT]) SetNext(next pkg.NodeInterface[KeyT, ValueT]) {
	n.Next = next.(*DataNode[KeyT, ValueT])
}

func NewDLList[KeyT comparable, ValueT any]() *pkg.DLList[KeyT, ValueT] {
	head := &DataNode[KeyT, ValueT]{}
	tail := &DataNode[KeyT, ValueT]{}
	head.SetNext(tail)
	tail.SetPrev(head)
	return &pkg.DLList[KeyT, ValueT]{Head: head, Tail: tail}
}

// Структура Cache: Описывает, что кэш использует хеш-таблицу для быстрого доступа к элементам и
// двусвязный список для отслеживания порядка использования элементов.
type Cache[KeyT comparable, ValueT any] struct {
	Capacity int
	Hash     map[KeyT]*DataNode[KeyT, ValueT]
	List     *pkg.DLList[KeyT, ValueT]
	Lock     sync.RWMutex
}

func NewCache[KeyT comparable, ValueT any](capacity int) *Cache[KeyT, ValueT] {
	return &Cache[KeyT, ValueT]{
		Capacity: capacity,
		Hash:     make(map[KeyT]*DataNode[KeyT, ValueT]),
		List:     NewDLList[KeyT, ValueT](),
	}
}

// Get извлекает значение из кэша по заданному ключу.
// Возвращает значение и true, если ключ найден, иначе возвращает нулевое значение и false.
func (c *Cache[KeyT, ValueT]) Get(key KeyT) (ValueT, bool) {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	node, ok := c.Hash[key]
	if ok {
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
		c.evictLocked()
	}

	// Создаем новый узел и добавляем его в кэш
	newNode := &DataNode[KeyT, ValueT]{
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
func (c *Cache[KeyT, ValueT]) evictLocked() {
	back := c.List.Back()
	if back != nil {
		c.List.Remove(back)
		delete(c.Hash, back.(*DataNode[KeyT, ValueT]).Key)
	}
}
