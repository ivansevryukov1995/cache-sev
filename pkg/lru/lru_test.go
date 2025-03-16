package lru

import (
	"sync"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	const cacheCapacity = 2
	const ttl = time.Millisecond * 100
	const cleanupInterval = time.Millisecond * 10

	cache := NewCache[int, string](cacheCapacity, ttl, cleanupInterval)

	// Проверка добавления и получения элемента
	cache.Put(1, "one")
	value, found := cache.Get(1)
	if !found || value != "one" {
		t.Errorf("Expected 'one', got '%v'", value)
	}

	// Проверка, что кэш возвращает false для отсутствующего ключа
	_, found = cache.Get(2)
	if found {
		t.Error("Expected not found for key 2")
	}

	// Проверка замены значения существующего ключа
	cache.Put(1, "uno")
	value, found = cache.Get(1)
	if !found || value != "uno" {
		t.Errorf("Expected 'uno', got '%v'", value)
	}

	// Проверка работы лимита емкости
	cache.Put(2, "two")
	cache.Put(3, "three") // Должен убрать ключ 1
	_, found = cache.Get(1)
	if found {
		t.Error("Expected key 1 to be evicted")
	}

	// Проверка, что ключ 2 все еще доступен
	value, found = cache.Get(2)
	if !found || value != "two" {
		t.Errorf("Expected 'two', got '%v'", value)
	}

	// Проверка, что кэш возвращает false для убранного ключа
	_, found = cache.Get(3)
	if !found {
		t.Error("Expected to find key 3")
	}

	// Должен вернуть значение 3
	value, found = cache.Get(3)
	if !found || value != "three" {
		t.Errorf("Expected 'three', got '%v'", value)
	}
}

func TestCacheRaceCondition(t *testing.T) {
	const numItems = 50
	const cacheCapacity = 100
	const ttl = time.Millisecond * 100
	const cleanupInterval = time.Millisecond * 10

	cache := NewCache[int, string](cacheCapacity, ttl, cleanupInterval)

	var wg sync.WaitGroup

	for i := 0; i < numItems; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			cache.Put(i, "value")
		}(i)

		go func(i int) {
			defer wg.Done()
			_, _ = cache.Get(i)
		}(i)
	}

	wg.Wait()

	for i := 0; i < numItems; i++ {
		value, found := cache.Get(i)
		if !found {
			t.Errorf("Expected to find key %d", i)
		} else if value != "value" {
			t.Errorf("Expected 'value', got '%v'", value)
		}
	}
}
