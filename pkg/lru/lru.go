package lru

import (
	"sync"
	"time"
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
	capacity      int
	cache         map[KeyT]*Node[KeyT, ValueT]
	list          *List[KeyT, ValueT]
	mutex         sync.RWMutex
	ttl           time.Duration
	cleanupTicker *time.Ticker  // Тикер для периодической очистки
	stopCleanup   chan struct{} // Канал для остановки процесса очистки
}

func NewList[KeyT comparable, ValueT any]() *List[KeyT, ValueT] {
	var key KeyT
	var value ValueT
	list := &List[KeyT, ValueT]{
		head: &Node[KeyT, ValueT]{key, value, nil, nil, time.Time{}},
		tail: &Node[KeyT, ValueT]{key, value, nil, nil, time.Time{}},
	}
	list.head.next = list.tail
	list.tail.prev = list.head
	return list
}

func NewCache[KeyT comparable, ValueT any](capacity int, ttl time.Duration, cleanupInterval time.Duration) *Cache[KeyT, ValueT] {
	cache := &Cache[KeyT, ValueT]{
		capacity:      capacity,
		cache:         make(map[KeyT]*Node[KeyT, ValueT]),
		list:          NewList[KeyT, ValueT](),
		ttl:           ttl,
		stopCleanup:   make(chan struct{}), // Инициализация канала для остановки очистки
		cleanupTicker: time.NewTicker(cleanupInterval),
	}

	go cache.cleanupExpired() // Запуск фонового процесса очистки

	return cache
}

// cleanupExpired проверяет и удаляет устаревшие элементы в кэше
func (cache *Cache[KeyT, ValueT]) cleanupExpired() {
	for {
		select {
		case <-cache.cleanupTicker.C:
			cache.mutex.Lock()
			for key, node := range cache.cache {
				if time.Now().After(node.expiry) {
					cache.Remove(key) // Удаление устаревшего элемента
				}
			}
			cache.mutex.Unlock()
		case <-cache.stopCleanup: // Получение сигнала для остановки
			return
		}
	}
}

func (l *List[KeyT, ValueT]) PushToFront(node *Node[KeyT, ValueT]) {
	node.prev = l.head
	node.next = l.head.next
	l.head.next.prev = node
	l.head.next = node
}

func (l *List[KeyT, ValueT]) Remove(node *Node[KeyT, ValueT]) {
	if node == nil {
		return
	}
	prev := node.prev
	next := node.next
	prev.next = next
	next.prev = prev
}

func (l *List[KeyT, ValueT]) MoveToFront(node *Node[KeyT, ValueT]) {
	if node == nil {
		return
	}
	l.Remove(node)
	l.PushToFront(node)
}

func (l *List[KeyT, ValueT]) Back() *Node[KeyT, ValueT] {
	if l.tail == l.head {
		return nil
	}
	return l.tail.prev
}

func (cache *Cache[KeyT, ValueT]) Get(key KeyT) (ValueT, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if node, found := cache.cache[key]; found {
		if time.Now().After(node.expiry) {
			cache.Remove(key)
			var value ValueT
			return value, false
		}
		cache.list.MoveToFront(node)
		return node.value, true
	}
	var value ValueT
	return value, false
}

func (cache *Cache[KeyT, ValueT]) Remove(key KeyT) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if node, found := cache.cache[key]; found {
		cache.list.Remove(node)
		delete(cache.cache, key)
	}
}

func (cache *Cache[KeyT, ValueT]) Put(key KeyT, value ValueT) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if node, found := cache.cache[key]; found {
		cache.list.MoveToFront(node)
		node.value = value
		node.expiry = time.Now().Add(cache.ttl)
		return
	}

	if len(cache.cache) == cache.capacity {
		back := cache.list.Back()
		if back != nil {
			cache.list.Remove(back)
			delete(cache.cache, back.key)
		}
	}

	newNode := &Node[KeyT, ValueT]{key, value, nil, nil, time.Now().Add(cache.ttl)}
	cache.list.PushToFront(newNode)
	cache.cache[key] = newNode
}

// StopCleanup останавливает процесс очистки
func (cache *Cache[KeyT, ValueT]) StopCleanup() {
	close(cache.stopCleanup)   // Посылаем сигнал для остановки
	cache.cleanupTicker.Stop() // Остановка тика
}
