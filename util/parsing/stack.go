package parsing

import (
	"slices"
)

type Stack[T TokenTyper] struct {
	elems  []*Token[T]
	popped []*Token[T]
}

func NewStack[T TokenTyper]() *Stack[T] {
	return &Stack[T]{
		elems:  make([]*Token[T], 0),
		popped: make([]*Token[T], 0),
	}
}

func (s *Stack[T]) Peek() (*Token[T], bool) {
	if len(s.elems) == 0 {
		return nil, false
	}

	return s.elems[len(s.elems)-1], true
}

func (s *Stack[T]) Push(tok *Token[T]) {
	if tok == nil {
		return
	}

	s.elems = append(s.elems, tok)
}

func (s *Stack[T]) Pop() (*Token[T], bool) {
	if len(s.elems) == 0 {
		return nil, false
	}

	tok := s.elems[len(s.elems)-1]
	s.elems = s.elems[:len(s.elems)-1]

	s.popped = append(s.popped, tok)

	return tok, true
}

func (s *Stack[T]) RefuseOne() bool {
	if len(s.popped) == 0 {
		return false
	}

	tok := s.popped[len(s.popped)-1]
	s.popped = s.popped[:len(s.popped)-1]

	s.elems = append(s.elems, tok)

	return true
}

func (s *Stack[T]) RefuseMany() {
	if len(s.popped) == 0 {
		return
	}

	for len(s.popped) > 0 {
		tok := s.popped[len(s.popped)-1]
		s.popped = s.popped[:len(s.popped)-1]

		s.elems = append(s.elems, tok)
	}
}

func (s *Stack[T]) Accept() {
	if len(s.popped) == 0 {
		return
	}

	s.popped = s.popped[:0]
}

func (s *Stack[T]) Size() int {
	return len(s.elems)
}

func (s *Stack[T]) GetPopped() []*Token[T] {
	if len(s.popped) == 0 {
		return nil
	}

	popped := make([]*Token[T], len(s.popped))
	copy(popped, s.popped)

	slices.Reverse(popped)

	return popped
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.elems) == 0
}
