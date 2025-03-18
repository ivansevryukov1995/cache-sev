package pkg

type Node[KeyT comparable, ValueT any] struct {
	Key   KeyT
	Value ValueT
	Prev  *Node[KeyT, ValueT]
	Next  *Node[KeyT, ValueT]
}
