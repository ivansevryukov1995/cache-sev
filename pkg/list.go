package pkg

type List[KeyT comparable, ValueT any] struct {
	Head *Node[KeyT, ValueT]
	Tail *Node[KeyT, ValueT]
}

func NewList[KeyT comparable, ValueT any]() *List[KeyT, ValueT] {
	head := &Node[KeyT, ValueT]{}
	tail := &Node[KeyT, ValueT]{}
	head.Next = tail
	tail.Prev = head
	return &List[KeyT, ValueT]{Head: head, Tail: tail}
}

func (l *List[KeyT, ValueT]) PushToFront(node *Node[KeyT, ValueT]) {
	node.Prev = l.Head
	node.Next = l.Head.Next
	l.Head.Next.Prev = node
	l.Head.Next = node
}

func (l *List[KeyT, ValueT]) Remove(node *Node[KeyT, ValueT]) {
	if node == nil {
		return
	}
	node.Prev.Next = node.Next
	node.Next.Prev = node.Prev
}

func (l *List[KeyT, ValueT]) MoveToFront(node *Node[KeyT, ValueT]) {
	if node == nil || node.Prev == nil {
		return
	}
	l.Remove(node)
	l.PushToFront(node)
}

func (l *List[KeyT, ValueT]) Back() *Node[KeyT, ValueT] {
	if l.Tail.Prev == l.Head {
		return nil
	}
	return l.Tail.Prev
}
