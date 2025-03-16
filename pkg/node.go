package pkg

import "time"

type Node[KeyT comparable, ValueT any] struct {
	Key       KeyT
	Value     ValueT
	Prev      *Node[KeyT, ValueT]
	Next      *Node[KeyT, ValueT]
	ExpiresAt time.Time // Время, когда узел срок действия истекает
}
