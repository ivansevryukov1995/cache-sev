package pkg

type DLList[KeyT comparable, ValueT any] struct {
	Head NodeInterface[KeyT, ValueT]
	Tail NodeInterface[KeyT, ValueT]
}

func (l *DLList[KeyT, ValueT]) PushToFront(node NodeInterface[KeyT, ValueT]) {
	node.SetPrev(l.Head)
	node.SetNext(l.Head.GetNext())
	l.Head.GetNext().SetPrev(node)
	l.Head.SetNext(node)
}

func (l *DLList[KeyT, ValueT]) Remove(node NodeInterface[KeyT, ValueT]) {
	if node == nil {
		return
	}

	if node == l.Head {
		l.Head = node.GetNext()
	}

	if node == l.Tail {
		l.Tail = node.GetPrev()
	}

	if node.GetPrev() != nil {
		node.GetPrev().SetNext(node.GetNext())
	}
	if node.GetNext() != nil {
		node.GetNext().SetPrev(node.GetPrev())
	}

	// Обнуляем указатели узла, чтобы избежать утечек памяти
	// node.SetPrev(nil)
	// node.SetNext(nil)
}

func (l *DLList[KeyT, ValueT]) MoveToFront(node NodeInterface[KeyT, ValueT]) {
	if node == nil || node.GetPrev() == nil {
		return
	}
	l.Remove(node)
	l.PushToFront(node)
}

func (l *DLList[KeyT, ValueT]) Back() NodeInterface[KeyT, ValueT] {
	if l.Tail.GetPrev() == l.Head {
		return nil
	}
	return l.Tail.GetPrev()
}
