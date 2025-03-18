package lru

import (
	"testing"
	"time"
)

// MockLogger - простая реализация интерфейса Logger для тестирования
type MockLogger struct {
	messages []string
}

func (ml *MockLogger) Log(message string) {
	ml.messages = append(ml.messages, message)
}

func TestCache(t *testing.T) {
	logger := &MockLogger{}
	cache := NewCache[string, string](2) // Кэш с максимальным размером 2 и TTL 2 секунды
	cache.Logger = logger

	// Тестирование добавления элементов
	cache.Put("key1", "value1")
	if val, found := cache.Get("key1"); !found || val != "value1" {
		t.Errorf("Expected value1, got %v (found: %v)", val, found)
	}

	// Тестирование обновления элемента
	cache.Put("key1", "value_updated")
	time.Sleep(100 * time.Millisecond) // небольшая задержка перед следующей операцией
	if val, found := cache.Get("key1"); !found || val != "value_updated" {
		t.Errorf("Expected value_updated, got %v (found: %v)", val, found)
	}

	// Тестирование истечения срока действия
	time.Sleep(3 * time.Second) // Ждем, пока TTL истечет
	if _, found := cache.Get("key1"); found {
		t.Error("Expected key1 to be expired")
	}

	// Тестирование добавления элемент, превышающего емкость
	// cache.Put("key1", "value1")
	cache.Put("key2", "value2")
	cache.Put("key3", "value3") // Должен удалить key1
	if _, found := cache.Get("key1"); found {
		t.Error("Expected key1 to be evicted")
	}
	if val, found := cache.Get("key3"); !found || val != "value3" {
		t.Errorf("Expected value3, got %v (found: %v)", val, found)
	}

	// Проверка логов
	expectedLogCount := 6 // Общее количество сообщений, ожидаемое с учетом каждого действия
	if len(logger.messages) < expectedLogCount {
		t.Errorf("Expected at least %d log messages, got %d", expectedLogCount, len(logger.messages))
	}
}

// Тест на удаление старого элемента
func TestCacheOldestEviction(t *testing.T) {
	logger := &MockLogger{}
	cache := NewCache[string, string](1) // Кэш с максимальным размером 1

	cache.Logger = logger
	cache.Put("key1", "value1")
	time.Sleep(100 * time.Millisecond) // небольшая задержка перед следующей операцией
	cache.Put("key2", "value2")        // Должен удалить key1
	if _, found := cache.Get("key1"); found {
		t.Error("Expected key1 to be evicted")
	}
	if val, found := cache.Get("key2"); !found || val != "value2" {
		t.Errorf("Expected value2, got %v (found: %v)", val, found)
	}

	// Проверка логов
	if len(logger.messages) < 2 {
		t.Errorf("Expected at least 2 log messages, got %d", len(logger.messages))
	}
}

// Тест на истечение срока действия
func TestCacheTTL(t *testing.T) {
	logger := &MockLogger{}
	cache := NewCache[string, string](2, 1*time.Second) // Кэш с максимальным размером 2 и TTL 1 секунда

	cache.Logger = logger
	cache.Put("key1", "value1")
	time.Sleep(2 * time.Second) // Ждем, пока TTL истечет
	if _, found := cache.Get("key1"); found {
		t.Error("Expected key1 to be expired")
	}

	// Проверка логов
	if len(logger.messages) < 1 {
		t.Errorf("Expected at least 1 log message, got %d", len(logger.messages))
	}
}
