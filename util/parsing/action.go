package parsing

import uc "github.com/PlayerR9/MyGoLib/Units/common"

type Actioner[T TokenTyper] interface {
	GetLHS() T

	uc.Iterable[T]
}

type Action[T TokenTyper] struct {
	rule *Rule[T]
}

func (a *Action[T]) GetLHS() T {
	return a.rule.lhs
}

func (a *Action[T]) Iterator() uc.Iterater[T] {
	return uc.NewSimpleIterator(a.rule.rhss)
}

type ActReduce[T TokenTyper] struct {
	*Action[T]
}

// nil if rhss is empty
func NewActReduce[T TokenTyper](rule *Rule[T]) *ActReduce[T] {
	if rule == nil {
		return nil
	}

	return &ActReduce[T]{
		Action: &Action[T]{
			rule: rule,
		},
	}
}

type ActShift[T TokenTyper] struct {
	*Action[T]
}

func NewActShift[T TokenTyper]() *ActShift[T] {
	return &ActShift[T]{
		Action: &Action[T]{
			rule: nil,
		},
	}
}

type ActAccept[T TokenTyper] struct {
	*Action[T]
}

// nil if rhss is empty
func NewActAccept[T TokenTyper](rule *Rule[T]) *ActAccept[T] {
	if rule == nil {
		return nil
	}

	return &ActAccept[T]{
		Action: &Action[T]{
			rule: rule,
		},
	}
}
