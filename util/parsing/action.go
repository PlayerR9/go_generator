package parsing

import (
	"fmt"

	uc "github.com/PlayerR9/lib_units/common"
)

// Actioner is an action.
type Actioner[T TokenTyper] interface {
	// GetLHS returns the left hand side.
	//
	// Returns:
	//   - T: The left hand side.
	GetLHS() T

	uc.Iterable[T]
	fmt.Stringer
}

// Action is an action.
type Action[T TokenTyper] struct {
	// rule is the rule.
	rule *Rule[T]
}

// GetLHS implements the Actioner interface.
func (a *Action[T]) GetLHS() T {
	return a.rule.lhs
}

// Iterator implements the Actioner interface.
func (a *Action[T]) Iterator() uc.Iterater[T] {
	return uc.NewSimpleIterator(a.rule.rhss)
}

// String implements the Actioner interface.
func (a *Action[T]) String() string {
	panic("this should never be called")
}

// ActReduce is a reduce action; which is a type of action.
type ActReduce[T TokenTyper] struct {
	*Action[T]
}

// String implements the fmt.Stringer interface.
func (a *ActReduce[T]) String() string {
	return "reduce"
}

// NewActReduce creates a new reduce action.
//
// Parameters:
//   - rule: The rule.
//
// Returns:
//   - *ActReduce: The new reduce action. Nil if the rule is nil.
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

// ActShift is a shift action; which is a type of action.
type ActShift[T TokenTyper] struct {
	*Action[T]
}

// String implements the fmt.Stringer interface.
func (a *ActShift[T]) String() string {
	return "shift"
}

// NewActShift creates a new shift action.
//
// Parameters:
//   - rule: The rule.
//
// Returns:
//   - *ActShift: The new shift action. Never returns nil.
func NewActShift[T TokenTyper]() *ActShift[T] {
	return &ActShift[T]{
		Action: &Action[T]{
			rule: nil,
		},
	}
}

// ActAccept is an accept action; which is a type of action.
type ActAccept[T TokenTyper] struct {
	*Action[T]
}

// String implements the fmt.Stringer interface.
func (a *ActAccept[T]) String() string {
	return "accept"
}

// NewActAccept creates a new accept action.
//
// Parameters:
//   - rule: The rule.
//
// Returns:
//   - *ActAccept: The new accept action. Nil if the rule is nil.
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
