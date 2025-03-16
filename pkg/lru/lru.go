package lru

import "github.com/ivansevryukov1995/cache-sev/pkg"

type Cache[KeyT comparable, ValueT any] struct {
	Capacity int
	Hash     map[KeyT]*pkg.Node[KeyT, ValueT]
	List     *pkg.List[KeyT, ValueT]
}

func NewCache[KeyT comparable, ValueT any](capacity int) *Cache[KeyT, ValueT] {
	return &Cache[KeyT, ValueT]{
		Capacity: capacity,
		Hash:     make(map[KeyT]*pkg.Node[KeyT, ValueT]),
		List:     pkg.NewList[KeyT, ValueT](),
	}
}

func (c *Cache[KeyT, ValueT]) Get(key KeyT) (ValueT, bool) {
	if node, found := c.Hash[key]; found {
		c.List.MoveToFront(node)
		return node.Value, true
	}
	var zeroValue ValueT
	return zeroValue, false
}

func (c *Cache[KeyT, ValueT]) Put(key KeyT, value ValueT) {
	if node, found := c.Hash[key]; found {
		c.List.MoveToFront(node)
		node.Value = value
		return
	}
	if len(c.Hash) == c.Capacity {
		back := c.List.Back()
		if back != nil {
			c.List.Remove(back)
			delete(c.Hash, back.Key)
		}
	}
	newNode := &pkg.Node[KeyT, ValueT]{
		Key:   key,
		Value: value,
		Prev:  nil,
		Next:  nil,
	}
	c.List.PushToFront(newNode)
	c.Hash[key] = newNode
}
