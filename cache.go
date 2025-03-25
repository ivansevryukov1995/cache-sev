package main

import (
	"fmt"
	"time"

	"github.com/ivansevryukov1995/cache-sev/pkg/lfu"
	"github.com/ivansevryukov1995/cache-sev/pkg/lru"
)

type Cacher[KeyT comparable, ValueT any] interface {
	Get(key KeyT) (ValueT, bool)
	Put(key KeyT, value ValueT, ttl time.Duration)
}

// Acceptable values of the politics argument field lru, lfu
func NewCache[KeyT comparable, ValueT any](politics string, capacity int) (Cacher[KeyT, ValueT], error) {
	switch politics {
	case "lru":
		return lru.NewCache[KeyT, ValueT](capacity), nil
	case "lfu":
		return lfu.NewCache[KeyT, ValueT](capacity), nil
	default:
		return nil, fmt.Errorf("Wrong politics type passed")
	}
}
