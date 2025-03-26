package lfu

import (
	"sync"
	"time"

	"github.com/ivansevryukov1995/cache-sev/pkg"
)

// Node представляет элемент в LFU кэше, который хранит данные и ссылку на его родительский узел частоты.
type DataNode[KeyT comparable, ValueT any] struct {
	Parent *FreqNode[KeyT, ValueT]
	Key    KeyT
	Value  ValueT
	Prev   *DataNode[KeyT, ValueT]
	Next   *DataNode[KeyT, ValueT]
}

type FreqNode[KeyT comparable, ValueT any] struct {
	Freq  int
	List  *pkg.DLList[KeyT, ValueT]
	Value ValueT
	Prev  *FreqNode[KeyT, ValueT]
	Next  *FreqNode[KeyT, ValueT]
}

// Cache представляет сам LFU кэш.
type Cache[KeyT comparable, ValueT any] struct {
	Capacity int
	Hash     map[KeyT]*DataNode[KeyT, ValueT]
	FreqHead *FreqNode[KeyT, ValueT]
	Lock     sync.RWMutex
}

// NewDataNode создает новый элемент LFU.
func NewDataNode[KeyT comparable, ValueT any](data ValueT, key KeyT, parent *FreqNode[KeyT, ValueT]) *DataNode[KeyT, ValueT] {
	return &DataNode[KeyT, ValueT]{
		Parent: parent,
		Key:    key,
		Value:  data,
	}
}

func NewDLList[KeyT comparable, ValueT any]() *pkg.DLList[KeyT, ValueT] {
	head := &DataNode[KeyT, ValueT]{}
	tail := &DataNode[KeyT, ValueT]{}
	head.SetNext(tail)
	tail.SetPrev(head)
	return &pkg.DLList[KeyT, ValueT]{Head: head, Tail: tail}
}

// NewFreqNode создает новый узел частоты с заданным значением.
func NewFreqNode[KeyT comparable, ValueT any]() *FreqNode[KeyT, ValueT] {
	return &FreqNode[KeyT, ValueT]{
		Freq: 0,
		List: NewDLList[KeyT, ValueT](),
		Prev: nil,
		Next: nil,
	}
}

// NewCache создает новый LFU кэш.
func NewCache[KeyT comparable, ValueT any](capacity int) *Cache[KeyT, ValueT] {
	head := NewFreqNode[KeyT, ValueT]()
	head.SetPrev(head)
	head.SetNext(head)
	return &Cache[KeyT, ValueT]{
		Capacity: capacity,
		Hash:     make(map[KeyT]*DataNode[KeyT, ValueT]),
		FreqHead: head,
	}
}

// GetNewFreqNode создает новый узел частоты с заданным значением и устанавливает ссылки на предыдущий и следующий узлы.
func GetNewFreqNode[KeyT comparable, ValueT any](value int, prev, next *FreqNode[KeyT, ValueT]) *FreqNode[KeyT, ValueT] {
	newNode := NewFreqNode[KeyT, ValueT]()
	newNode.Freq = value
	newNode.SetPrev(prev)
	newNode.SetNext(next)
	prev.SetNext(newNode)
	next.SetPrev(newNode)
	return newNode
}

// DeleteFreqNode удаляет узел из списка.
func DeleteFreqNode[KeyT comparable, ValueT any](node *FreqNode[KeyT, ValueT]) {
	next := node.GetNext()
	prev := node.GetPrev()
	node.GetPrev().SetNext(next)
	node.GetNext().SetPrev(prev)
}

func (n *DataNode[KeyT, ValueT]) GetPrev() pkg.NodeInterface[KeyT, ValueT] {
	return n.Prev
}

func (n *DataNode[KeyT, ValueT]) GetNext() pkg.NodeInterface[KeyT, ValueT] {
	return n.Next
}

func (n *DataNode[KeyT, ValueT]) SetPrev(prev pkg.NodeInterface[KeyT, ValueT]) {
	n.Prev = prev.(*DataNode[KeyT, ValueT])
}

func (n *DataNode[KeyT, ValueT]) SetNext(next pkg.NodeInterface[KeyT, ValueT]) {
	n.Next = next.(*DataNode[KeyT, ValueT])
}

func (n *FreqNode[KeyT, ValueT]) GetPrev() pkg.NodeInterface[KeyT, ValueT] {
	return n.Prev
}

func (n *FreqNode[KeyT, ValueT]) GetNext() pkg.NodeInterface[KeyT, ValueT] {
	return n.Next
}

func (n *FreqNode[KeyT, ValueT]) SetPrev(prev pkg.NodeInterface[KeyT, ValueT]) {
	n.Prev = prev.(*FreqNode[KeyT, ValueT])
}

func (n *FreqNode[KeyT, ValueT]) SetNext(next pkg.NodeInterface[KeyT, ValueT]) {
	n.Next = next.(*FreqNode[KeyT, ValueT])
}

// При повторном обращении к этому элементу ищется узел частоты элемента и запрашивается значение его следующего брата.
// Если его брат не существует или значение его следующего брата не на 1 больше его значения,
// то создаётся новый узел частоты со значением на 1 больше значения этого узла частоты и вставляется на нужное место.
// Узел удаляется из текущего набора и вставляется в набор нового списка частот. Указатель частоты узла обновляется и указывает на новый узел с частотой.
// Например, если к узлу z обращаются ещё раз (1), то он удаляется из списка частот со значением 2 и добавляется в список частот со значением 3 (2).
// Таким образом, временная сложность доступа к элементу составляет O(1).
// Get получает элемент из кэша и увеличивает его счетчик использования.
func (c *Cache[KeyT, ValueT]) Get(key KeyT) (ValueT, bool) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	item, ok := c.Hash[key]
	if ok {
		value := item.Value
		c.updateLocked(key, value)

		return item.Value, true
	}

	var zeroValue ValueT

	return zeroValue, false
}

func (c *Cache[KeyT, ValueT]) Put(key KeyT, value ValueT, ttl time.Duration) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	if _, ok := c.Hash[key]; ok {
		c.updateLocked(key, value)
		return
	}

	if len(c.Hash) >= c.Capacity {
		c.evictLocked()
	}

	// Если следующая частота после частотной головы не равна 1,
	// создаем новый узел частоты
	freq := c.FreqHead.Next

	if freq.Freq != 1 {
		freq = GetNewFreqNode(1, c.FreqHead, freq)
	}

	// Создаем новый элемент и вставляем его
	// в начало двусвязного списка данной частоты,
	// добавляем в хеш-таблицу
	newNode := NewDataNode(value, key, freq)
	freq.List.PushToFront(newNode)
	c.Hash[key] = newNode

	if ttl > 0 {
		go func() {
			<-time.After(ttl)
			c.Lock.Lock()
			defer c.Lock.Unlock()
			if _, exists := c.Hash[key]; exists {
				parent := newNode.Parent
				parent.List.Remove(newNode)
				delete(c.Hash, key)
				// Удаляем родительский узел частоты
				// если двусвязного списка частоты пуст
				if parent.List.Head.GetNext() == parent.List.Tail {
					DeleteFreqNode(parent)
				}
			}
		}()
	}
}

// updateLocked обновляет частоту использования элемента
func (c *Cache[KeyT, ValueT]) updateLocked(key KeyT, value ValueT) {
	item, _ := c.Hash[key]

	item.Value = value

	// Если следующий узел частоты не существует
	// или его частота не на 1 больше, создаем новый узел
	freqParent := item.Parent
	nextFreq := freqParent.Next

	if nextFreq == c.FreqHead || nextFreq.Freq != freqParent.Freq+1 {
		nextFreq = GetNewFreqNode(freqParent.Freq+1, freqParent, nextFreq)
	}

	// Обновляем ссылку на родительский узел частоты
	item.Parent = nextFreq

	// Удаляем элемент из двусвязного списка прошлой частоты
	// Вставляем элемент, полученный по ключю,
	// в начало двусвязного списка данной частоты
	freqParent.List.Remove(item)
	nextFreq.List.PushToFront(item)

	// Удаляем родительский узел частоты
	// если двусвязного списка частоты пуст
	if freqParent.List.Head.GetNext() == freqParent.List.Tail {
		DeleteFreqNode(freqParent)
	}
}

// Наименее часто использовавшиеся (Least Frequently Used — LFU):
// убирает запись, которая использовалась наименее часто
func (c *Cache[KeyT, ValueT]) evictLocked() {
	minFreqNode := c.FreqHead.Next
	if minFreqNode == c.FreqHead {
		panic("No item to evict")
	}

	back := minFreqNode.List.Back()

	if back != nil {
		minFreqNode.List.Remove(back)
		key := back.(*DataNode[KeyT, ValueT]).Key
		delete(c.Hash, key)

		// Удаляем родительский узел частоты
		// если двусвязного списка частоты пуст
		if minFreqNode.List.Head.GetNext() == minFreqNode.List.Tail {
			DeleteFreqNode(minFreqNode)
		}
	}
}
