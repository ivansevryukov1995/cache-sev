package main

import (
	"fmt"
	"time"

	"github.com/ivansevryukov1995/cache-sev/pkg"
	"github.com/ivansevryukov1995/cache-sev/pkg/lfu"
	"github.com/ivansevryukov1995/cache-sev/pkg/lru"
)

type Cacher[KeyT comparable, ValueT any] interface {
	Get(key KeyT) (ValueT, bool)
	Put(key KeyT, value ValueT, ttl time.Duration)
	SetLogger(pkg.Logger)
}

func NewCache[KeyT comparable, ValueT any](politics string, capacity int) (Cacher[KeyT, ValueT], error) {
	if politics == "lru" {

		return lru.NewCache[KeyT, ValueT](capacity), nil
	}

	if politics == "lfu" {
		return lfu.NewCache[KeyT, ValueT](capacity), nil
	}

	return nil, fmt.Errorf("Wrong politics type passed")
}
