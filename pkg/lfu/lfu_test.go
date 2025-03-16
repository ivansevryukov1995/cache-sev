package lfu

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCacheRaceCondition(t *testing.T) {
	const numItems = 50
	const cacheCapacity = 100
	const ttl = time.Millisecond * 100
	const cleanupInterval = time.Millisecond * 10

	cache := NewCache[int, string](cacheCapacity, ttl, cleanupInterval)

	var wg sync.WaitGroup

	// Запускаем несколько горутин, которые одновременно добавляют элементы в кэш и пытаются их получить
	for i := 0; i < numItems; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			cache.Put(i, fmt.Sprintf("value%d", i))
		}(i)

		go func(i int) {
			defer wg.Done()
			_, _ = cache.Get(i) // Попытки получить значения
		}(i)
	}

	wg.Wait() // Дождаться завершения всех горутин

	// Проверка, что все элементы находятся в кэше
	for i := 0; i < numItems; i++ {
		value, found := cache.Get(i)
		if found && value != fmt.Sprintf("value%d", i) {
			t.Errorf("Expected value for key %d to be 'value%d', got '%v'", i, i, value)
		}
	}
	cache.base.StopCleanup() // Правильная остановка фонового процесса
}

func TestCachePutAndGet(t *testing.T) {
	const cacheCapacity = 2
	const ttl = time.Millisecond * 100
	const cleanupInterval = time.Millisecond * 10

	cache := NewCache[int, string](cacheCapacity, ttl, cleanupInterval) // Создаем кэш с емкостью 2

	cache.Put(1, "value1")
	cache.Put(2, "value2")

	// Проверяем, что мы можем получить сохраненные значения
	if v, found := cache.Get(1); !found || v != "value1" {
		t.Errorf("Expected to find key 1, got %v", v)
	}
	if v, found := cache.Get(2); !found || v != "value2" {
		t.Errorf("Expected to find key 2, got %v", v)
	}

	// Добавляем третий элемент, который должен вызвать вытеснение
	cache.Put(3, "value3")

	// Ключ 1 должен быть вытеснен, поэтому мы не должны его найти
	if _, found := cache.Get(1); found {
		t.Errorf("Expected to not find key 1 after eviction")
	}

	// Ключ 2 должен быть доступен
	if v, found := cache.Get(2); !found || v != "value2" {
		t.Errorf("Expected to find key 2, got %v", v)
	}

	// Проверяем, что ключ 3 доступен
	if v, found := cache.Get(3); !found || v != "value3" {
		t.Errorf("Expected to find key 3, got %v", v)
	}

	cache.base.StopCleanup() // Правильная остановка фонового процесса
}

func TestCacheEviction(t *testing.T) {
	const cacheCapacity = 2
	const ttl = time.Millisecond * 100
	const cleanupInterval = time.Millisecond * 10

	cache := NewCache[int, string](cacheCapacity, ttl, cleanupInterval) // Создаем кэш с емкостью 2

	cache.Put(1, "value1")
	cache.Put(2, "value2")
	cache.Put(3, "value3") // Должен вытеснить key 1, так как они использовались с одинаковой частотой

	if _, found := cache.Get(1); found {
		t.Errorf("Expected not to find key 1 after eviction")
	}

	if v, found := cache.Get(2); !found || v != "value2" {
		t.Errorf("Expected to find key 2, got %v", v)
	}

	if v, found := cache.Get(3); !found || v != "value3" {
		t.Errorf("Expected to find key 3, got %v", v)
	}

	// Запускаем дополнительные добавления, чтобы проверить, что кэш правильно работает
	cache.Put(4, "value4") // Должен вытеснить key 2 (наименее часто используемый элемент)

	if _, found := cache.Get(2); found {
		t.Errorf("Expected not to find key 2 after eviction")
	}

	if v, found := cache.Get(3); !found || v != "value3" {
		t.Errorf("Expected to find key 3, got %v", v)
	}

	if v, found := cache.Get(4); !found || v != "value4" {
		t.Errorf("Expected to find key 4, got %v", v)
	}
	cache.base.StopCleanup()
}
