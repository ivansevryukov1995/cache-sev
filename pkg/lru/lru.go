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
	capacity int
	cache    map[KeyT]*Node[KeyT, ValueT]
	list     *List[KeyT, ValueT]
	mutex    sync.Mutex
	ttl      time.Duration
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

func NewCache[KeyT comparable, ValueT any](capacity int, ttl time.Duration) *Cache[KeyT, ValueT] {
	return &Cache[KeyT, ValueT]{
		capacity: capacity,
		cache:    make(map[KeyT]*Node[KeyT, ValueT]),
		list:     NewList[KeyT, ValueT](),
		ttl:      ttl,
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
