package lfu

import (
	"testing"
	"time"
)

func TestCachePutAndGet(t *testing.T) {
	const cacheCapacity = 2
	const ttl = time.Millisecond * 100

	cache := NewCache[int, string](cacheCapacity, ttl)

	cache.Put(1, "value1")
	cache.Put(2, "value2")

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

}

func TestCacheEviction(t *testing.T) {
	const cacheCapacity = 2
	const ttl = time.Millisecond * 200

	cache := NewCache[int, string](cacheCapacity, ttl)

	cache.Put(1, "value1")
	cache.Put(2, "value2")

	cache.Put(3, "value3") // Вытеснит один элемент из кэша

	// Ключ 1 должен быть вытеснен, потому что он использовался реже
	if _, found := cache.Get(1); found {
		t.Errorf("Expected to not find key 1 after eviction")
	}

	// Проверяем, что ключи 2 и 3 доступны
	if v, found := cache.Get(2); !found || v != "value2" {
		t.Errorf("Expected to find key 2, got %v", v)
	}
	if v, found := cache.Get(3); !found || v != "value3" {
		t.Errorf("Expected to find key 3, got %v", v)
	}

}

func TestCacheTTLExpiration(t *testing.T) {
	const cacheCapacity = 2
	const ttl = time.Millisecond * 100

	cache := NewCache[int, string](cacheCapacity, ttl)

	cache.Put(1, "value1")
	// Подождем, чтобы считываемое значение истекло
	time.Sleep(ttl + time.Millisecond*10)

	if _, found := cache.Get(1); found {
		t.Errorf("Expected to not find key 1 after TTL expiration")
	}

}

func TestCacheCleanup(t *testing.T) {
	const cacheCapacity = 2
	const ttl = time.Millisecond * 100

	cache := NewCache[int, string](cacheCapacity, ttl)

	cache.Put(1, "value1")
	cache.Put(2, "value2")
	// Подождем, чтобы ключи истекли
	time.Sleep(ttl + time.Millisecond*10)

	// Запускаем очистку

	if _, found := cache.Get(1); found {
		t.Errorf("Expected to not find key 1 after TTL expiration")
	}
	if _, found := cache.Get(2); found {
		t.Errorf("Expected to not find key 2 after TTL expiration")
	}
}
