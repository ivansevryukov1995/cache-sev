package lru

import (
	"time"

	"github.com/ivansevryukov1995/cache-sev/pkg/cache"
)

type Node[KeyT comparable, ValueT any] struct {
	key    KeyT
	value  ValueT
	prev   *Node[KeyT, ValueT]
	next   *Node[KeyT, ValueT]
	expiry time.Time
}

type List[KeyT comparable, ValueT any] struct {
	head *Node[KeyT, ValueT]
	tail *Node[KeyT, ValueT]
}

type Cache[KeyT comparable, ValueT any] struct {
	base     *cache.BaseCache[KeyT, ValueT]
	capacity int
	cache    map[KeyT]*Node[KeyT, ValueT]
	list     *List[KeyT, ValueT]
}

// NewList создает новый двусвязный список
func NewList[KeyT comparable, ValueT any]() *List[KeyT, ValueT] {
	list := &List[KeyT, ValueT]{
		head: &Node[KeyT, ValueT]{},
		tail: &Node[KeyT, ValueT]{},
	}
	list.head.next = list.tail
	list.tail.prev = list.head
	return list
}

// NewCache создает новый кэш с заданной емкостью и временем жизни
func NewCache[KeyT comparable, ValueT any](capacity int, ttl time.Duration, cleanupInterval time.Duration) *Cache[KeyT, ValueT] {
	base := cache.NewBaseCache[KeyT, ValueT](ttl, cleanupInterval)
	return &Cache[KeyT, ValueT]{
		base:     base,
		capacity: capacity,
		cache:    make(map[KeyT]*Node[KeyT, ValueT]),
		list:     NewList[KeyT, ValueT](),
	}
}

// PushToFront добавляет узел в начало списка
func (l *List[KeyT, ValueT]) PushToFront(node *Node[KeyT, ValueT]) {
	node.prev = l.head
	node.next = l.head.next
	l.head.next.prev = node
	l.head.next = node
}

// Remove удаляет узел из списка
func (l *List[KeyT, ValueT]) Remove(node *Node[KeyT, ValueT]) {
	if node == nil {
		return
	}
	node.prev.next = node.next
	node.next.prev = node.prev
}

// MoveToFront перемещает узел в начало списка
func (l *List[KeyT, ValueT]) MoveToFront(node *Node[KeyT, ValueT]) {
	l.Remove(node)
	l.PushToFront(node)
}

// Back возвращает последний узел списка
func (l *List[KeyT, ValueT]) Back() *Node[KeyT, ValueT] {
	if l.tail.prev == l.head {
		return nil // Список пуст
	}
	return l.tail.prev
}

// Get возвращает значение по ключу и проверяет его срок действия
func (cache *Cache[KeyT, ValueT]) Get(key KeyT) (ValueT, bool) {
	cache.base.Mutex.RLock()
	defer cache.base.Mutex.RUnlock()

	if node, found := cache.cache[key]; found {
		if time.Now().After(node.expiry) {
			cache.Remove(key)
			var zeroValue ValueT
			return zeroValue, false
		}
		cache.list.MoveToFront(node)
		return node.value, true
	}
	var zeroValue ValueT
	return zeroValue, false
}

// Remove удаляет элемент по ключу
func (cache *Cache[KeyT, ValueT]) Remove(key KeyT) {
	cache.base.Mutex.RLock()
	defer cache.base.Mutex.RUnlock()

	if node, found := cache.cache[key]; found {
		cache.list.Remove(node)
		delete(cache.cache, key)
	}
}

// Put добавляет новый элемент в кэш
func (cache *Cache[KeyT, ValueT]) Put(key KeyT, value ValueT) {
	cache.base.Mutex.RLock()
	defer cache.base.Mutex.RUnlock()

	if node, found := cache.cache[key]; found {
		// Если элемент уже существует, обновляем его
		cache.list.MoveToFront(node)
		node.value = value
		node.expiry = time.Now().Add(cache.base.Ttl)
		return
	}

	// Если кэш переполнен, удаляем последний элемент
	if len(cache.cache) >= cache.capacity {
		back := cache.list.Back()
		if back != nil {
			cache.Remove(back.key)
		}
	}

	newNode := &Node[KeyT, ValueT]{key, value, nil, nil, time.Now().Add(cache.base.Ttl)}
	cache.list.PushToFront(newNode)
	cache.cache[key] = newNode
}

// cleanupExpired периодически удаляет просроченные элементы
func (cache *Cache[KeyT, ValueT]) cleanupExpired() {
	for {
		select {
		case <-cache.base.CleanupTicker.C:
			now := time.Now()
			cache.base.Mutex.Lock() // Используйте Lock, так как мы изменяем состояние кэша
			for key, node := range cache.cache {
				if now.After(node.expiry) {
					cache.Remove(key)
				}
			}
			cache.base.Mutex.Unlock()
		case <-cache.base.StopCleanUp:
			return
		}
	}
}
