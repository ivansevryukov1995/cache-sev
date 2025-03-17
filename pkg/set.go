package pkg

import "fmt"

type Set[T comparable] struct {
	Items map[T]struct{}
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		Items: make(map[T]struct{}),
	}
}

func (s *Set[T]) Add(val T) {
	s.Items[val] = struct{}{}
}

func (s *Set[T]) Remove(val T) {
	delete(s.Items, val)
}

func (s *Set[T]) Pop() T {
	if len(s.Items) == 0 {
		panic("Pop() from the empty Set")
	}
	for key := range s.Items {
		s.Remove(key)
		return key
	}
	var ans T
	return ans
}

func (s *Set[T]) IsEmpty() bool {
	return len(s.Items) == 0
}

func (s *Set[T]) String() string {
	return fmt.Sprintf("%v", s.Items)
}
