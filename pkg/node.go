package pkg

type NodeInterface[KeyT comparable, ValueT any] interface {
	GetPrev() NodeInterface[KeyT, ValueT]
	GetNext() NodeInterface[KeyT, ValueT]
	SetPrev(NodeInterface[KeyT, ValueT])
	SetNext(NodeInterface[KeyT, ValueT])
}
