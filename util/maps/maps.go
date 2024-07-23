package maps

type SeenMap[T comparable] struct {
	table map[T]bool
}

func NewSeenMap[T comparable]() *SeenMap[T] {
	return &SeenMap[T]{
		table: make(map[T]bool),
	}
}

func (s *SeenMap[T]) See(key T) {
	s.table[key] = true
}

func (s *SeenMap[T]) IsSeen(key T) bool {
	v, ok := s.table[key]
	return ok && v
}

func (s *SeenMap[T]) FilterSeen(elems []T) []T {
	var top int

	for i := 0; i < len(elems); i++ {
		elem := elems[i]

		if !s.IsSeen(elem) {
			elems[top] = elems[i]
			top++
		}
	}

	return elems[:top]
}
