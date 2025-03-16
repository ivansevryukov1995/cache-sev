package pkg

import (
	"sync"
	"time"
)

// BaseCache содержит общие поля и методы для очистки
type BaseCache[KeyT comparable, ValueT any] struct {
	Ttl           time.Duration
	CleanupTicker *time.Ticker
	StopCleanUp   chan struct{}
	Mutex         sync.RWMutex
}

// NewBaseCache инициализирует базовую структуру для кэша
func NewBaseCache[KeyT comparable, ValueT any](ttl time.Duration, cleanupInterval time.Duration) *BaseCache[KeyT, ValueT] {
	base := &BaseCache[KeyT, ValueT]{
		Ttl:           ttl,
		StopCleanUp:   make(chan struct{}),
		CleanupTicker: time.NewTicker(cleanupInterval),
	}
	go base.cleanupExpired()
	return base
}

// cleanupExpired должна быть переопределена в конкретных реализациях
func (b *BaseCache[KeyT, ValueT]) cleanupExpired() {
	// Логика очистки должна быть оформлена в потомках
}

// StopCleanup останавливает процесс очистки
func (b *BaseCache[KeyT, ValueT]) StopCleanup() {
	close(b.StopCleanUp)
	b.CleanupTicker.Stop()
}
