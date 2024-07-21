package parsing

import uc "github.com/PlayerR9/MyGoLib/Units/common"

type Actioner[T uc.Enumer] interface {
	GetLHS() T

	uc.Iterable[T]
}

type Action[T uc.Enumer] struct {
	lhs  T
	rhss []T
}

func (a *Action[T]) GetLHS() T {
	return a.lhs
}

func (a *Action[T]) Iterator() uc.Iterater[T] {
	return uc.NewSimpleIterator(a.rhss)
}

type ActReduce[T uc.Enumer] struct {
	*Action[T]
}

// nil if rhss is empty
func NewActReduce[T uc.Enumer](lhs T, rhss []T) *ActReduce[T] {
	if len(rhss) == 0 {
		return nil
	}

	return &ActReduce[T]{
		Action: &Action[T]{
			lhs:  lhs,
			rhss: rhss,
		},
	}
}

type ActShift[T uc.Enumer] struct {
	*Action[T]
}

func NewActShift[T uc.Enumer]() *ActShift[T] {
	return &ActShift[T]{
		Action: &Action[T]{
			rhss: make([]T, 0),
		},
	}
}

type ActAccept[T uc.Enumer] struct {
	*Action[T]
}

// nil if rhss is empty
func NewActAccept[T uc.Enumer](lhs T, rhss []T) *ActAccept[T] {
	if len(rhss) == 0 {
		return nil
	}

	return &ActAccept[T]{
		Action: &Action[T]{
			lhs:  lhs,
			rhss: rhss,
		},
	}
}
