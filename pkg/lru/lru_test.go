package lru

import (
	"testing"
	"time"
)

func TestCache(t *testing.T) {

	cache := NewCache[string, string](2) // Кэш с максимальным размером 2 и TTL 2 секунды

	// Тестирование добавления элементов
	cache.Put("key1", "value1", time.Second*0)
	if val, found := cache.Get("key1"); !found || val != "value1" {
		t.Errorf("Expected value1, got %v (found: %v)", val, found)
	}

	// Тестирование обновления элемента
	cache.Put("key1", "value_updated", time.Second*0)
	time.Sleep(100 * time.Millisecond) // небольшая задержка перед следующей операцией
	if val, found := cache.Get("key1"); !found || val != "value_updated" {
		t.Errorf("Expected value_updated, got %v (found: %v)", val, found)
	}

	// Тестирование добавления элемент, превышающего емкость
	// cache.Put("key1", "value1")
	cache.Put("key2", "value2", time.Second*0)
	cache.Put("key3", "value3", time.Second*0) // Должен удалить key1
	if _, found := cache.Get("key1"); found {
		t.Error("Expected key1 to be evicted")
	}
	if val, found := cache.Get("key3"); !found || val != "value3" {
		t.Errorf("Expected value3, got %v (found: %v)", val, found)
	}
}

// Тест на удаление старого элемента
func TestCacheOldestEviction(t *testing.T) {

	cache := NewCache[string, string](1) // Кэш с максимальным размером 1

	cache.Put("key1", "value1", time.Second*0)
	time.Sleep(100 * time.Millisecond)         // небольшая задержка перед следующей операцией
	cache.Put("key2", "value2", time.Second*0) // Должен удалить key1
	if _, found := cache.Get("key1"); found {
		t.Error("Expected key1 to be evicted")
	}
	if val, found := cache.Get("key2"); !found || val != "value2" {
		t.Errorf("Expected value2, got %v (found: %v)", val, found)
	}

}

// Тест на истечение срока действия
// func TestCacheTTL(t *testing.T) {

// 	cache := NewCache[string, string](2) // Кэш с максимальным размером 2 и TTL 1 секунда

// 	cache.Put("key1", "value1", time.Second*0)
// 	time.Sleep(2 * time.Second) // Ждем, пока TTL истечет
// 	if _, found := cache.Get("key1"); found {
// 		t.Error("Expected key1 to be expired")
// 	}

// }
