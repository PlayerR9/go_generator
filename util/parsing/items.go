package parsing

import (
	"fmt"
	"slices"
	"strings"

	uc "github.com/PlayerR9/MyGoLib/Units/common"
	utstr "github.com/PlayerR9/go_generator/util/strings"
)

// StringToTypeFunc is a function that transforms a string into a token type.
//
// Parameters:
//   - field: The field string.
//
// Returns:
//   - T: The token type.
//   - bool: True if the field is valid. False otherwise.
type StringToTypeFunc[T TokenTyper] func(field string) (T, bool)

// Rule is a rule.
type Rule[T TokenTyper] struct {
	// lhs is the left hand side.
	lhs T

	// rhss are the right hand sides.
	rhss []T
}

// String implements the fmt.Stringer interface.
func (r *Rule[T]) String() string {
	values := make([]string, 0, len(r.rhss))

	for _, rhs := range r.rhss {
		values = append(values, rhs.String())
	}

	var builder strings.Builder

	builder.WriteString(r.lhs.String())
	builder.WriteString(" = ")
	builder.WriteString(strings.Join(values, " "))
	builder.WriteString(" .")

	return builder.String()
}

// Iterator implements the uc.Iterable interface.
//
// Never returns nil
func (r *Rule[T]) Iterator() uc.Iterater[T] {
	return uc.NewSimpleIterator(r.rhss)
}

// Copy implements the uc.Copier interface.
func (r *Rule[T]) Copy() uc.Copier {
	rhss_copy := make([]T, len(r.rhss))
	copy(rhss_copy, r.rhss)

	return &Rule[T]{
		lhs:  r.lhs,
		rhss: rhss_copy,
	}
}

// NewRule creates a new rule.
//
// Parameters:
//   - lhs: The left hand side.
//   - rhss: The right hand sides.
//
// Returns:
//   - *Rule: The new rule. Nil if an error occurs.
func NewRule[T TokenTyper](lhs T, rhss []T) *Rule[T] {
	if len(rhss) == 0 {
		return nil
	}

	return &Rule[T]{
		lhs:  lhs,
		rhss: rhss,
	}
}

// GetRhsAt returns the right hand side at the given index.
//
// Parameters:
//   - idx: The index.
//
// Returns:
//   - T: The right hand side.
//   - bool: True if the index is valid. False otherwise.
func (r *Rule[T]) GetRhsAt(idx int) (T, bool) {
	if idx < 0 || idx >= len(r.rhss) {
		return *new(T), false
	}

	return r.rhss[idx], true
}

// GetLhs returns the left hand side.
//
// Returns:
//   - T: The left hand side.
func (r *Rule[T]) GetLhs() T {
	return r.lhs
}

// GetIndicesOfRhs returns the indices of the right hand side.
//
// Parameters:
//   - rhs: The right hand side to search.
//
// Returns:
//   - []int: The indices of the right hand side.
func (r *Rule[T]) GetIndicesOfRhs(rhs T) []int {
	var indices []int

	for i := 0; i < len(r.rhss); i++ {
		if r.rhss[i] == rhs {
			indices = append(indices, i)
		}
	}

	return indices
}

// GetRhsFromStart returns the right hand side from the start.
//
// Parameters:
//   - start: The start index.
//
// Returns:
//   - T: The right hand side.
//   - bool: True if the start is valid. False otherwise.
func (r *Rule[T]) GetRhsFromStart(start int) (T, bool) {
	if start < 0 || start >= len(r.rhss) {
		return *new(T), false
	}

	return r.rhss[len(r.rhss)-start-1], true
}

// Item is an item in the rule.
type Item[T TokenTyper] struct {
	// rule is the rule. This contains a "copied" version of the rule. This can be modified.
	rule *Rule[T]

	// action is the action. This contains the original rule. Do not modify it.
	action Actioner[T]

	// pos is the position in the rule.
	pos int

	// lookaheads are the lookaheads.
	lookaheads []T
}

// String implements the fmt.Stringer interface.
func (item *Item[T]) String() string {
	uc.Assert(item.rule != nil, "item.rule must not be nil")

	var values []string

	// rule : ( action )
	iter := item.rule.Iterator()
	var i int

	for {
		rhs, err := iter.Consume()
		if err != nil {
			break
		}

		rhs_str := utstr.ConnectWords(rhs.String())

		if i == item.pos {
			values = append(values, "[")
			values = append(values, rhs_str)
			values = append(values, "]")
		} else {
			values = append(values, rhs_str)
		}

		i++
	}

	var act_str string

	if item.action != nil {
		act_str = item.action.String()
	} else {
		act_str = "no action"
	}

	values = append(values, "->", utstr.ConnectWords(item.GetLhs().String()), ":", "(", act_str, ")", ".")

	return strings.Join(values, " ")
}

// NewItem creates a new item.
//
// Parameters:
//   - rule: The rule.
//   - pos: The position in the rule.
//   - act: The action.
//
// Returns:
//   - *Item: The new item.
//   - error: An error of type *common.ErrInvalidParameter if the position is invalid or the rule is nil.
func NewItem[T TokenTyper](rule *Rule[T], pos int, act Actioner[T]) (*Item[T], error) {
	if rule == nil {
		return nil, uc.NewErrNilParameter("rule")
	} else if pos < 0 || pos >= len(rule.rhss) {
		return nil, uc.NewErrInvalidParameter("pos", uc.NewErrOutOfBounds(pos, 0, len(rule.rhss)))
	}

	if act == nil {
		if pos > 0 {
			act = NewActShift[T]()
		} else {
			rhs, ok := rule.GetRhsAt(pos)
			uc.AssertOk(ok, "rule.GetRhsAt(%d)", pos)

			if rhs.IsAcceptSymbol() {
				act = NewActAccept(rule)
			} else {
				act = NewActReduce(rule)
			}
		}
	}

	r_copy, ok := rule.Copy().(*Rule[T])
	uc.AssertOk(ok, "rule.Copy() does not return a *Rule[T]")

	return &Item[T]{
		rule:   r_copy,
		pos:    pos,
		action: act,
	}, nil
}

// MatchLookahead matches the lookahead.
//
// Parameters:
//   - la: The lookahead.
//
// Returns:
//   - bool: True if the lookahead matches. False otherwise.
//   - error: An error if the lookahead does not match the rule.
//
// As a special case, if item does not need lookahead, nil is returned regardless of the lookahead.
func (item *Item[T]) MatchLookahead(la *T) (bool, error) {
	if len(item.lookaheads) == 0 {
		return false, nil
	}

	if la == nil {
		return false, fmt.Errorf("expected %q, got nothing instead", item.lookaheads[0].String())
	}

	_, ok := slices.BinarySearch(item.lookaheads, *la)
	if ok {
		return true, nil
	}

	values := make([]string, 0, len(item.lookaheads))

	for _, lookahead := range item.lookaheads {
		values = append(values, lookahead.String())
	}

	return false, fmt.Errorf("expected %s, got %q instead", uc.OrQuoteString(values, false), *la)
}

// GetLhs returns the left hand side.
//
// Returns:
//   - T: The left hand side.
func (item *Item[T]) GetLhs() T {
	uc.Assert(item.rule != nil, "item.rule must not be nil")

	return item.rule.GetLhs()
}

// GetRhsRelative returns the right hand side.
//
// Parameters:
//   - delta: The delta.
//
// Returns:
//   - T: The right hand side.
//   - bool: True if the index is valid. False otherwise.
func (item *Item[T]) GetRhsRelative(delta int) (T, bool) {
	uc.Assert(item.rule != nil, "item.rule must not be nil")

	pos := item.pos + delta

	if pos < 0 || pos >= len(item.rule.rhss) {
		return *new(T), false
	}

	return item.rule.rhss[pos], true
}

// IsDone returns true if the item is done.
//
// Parameters:
//   - delta: The delta.
//
// Returns:
//   - bool: True if the item is done. False otherwise.
func (item *Item[T]) IsDone(delta int) bool {
	uc.Assert(item.rule != nil, "item.rule must not be nil")

	pos := item.pos + delta

	return pos >= len(item.rule.rhss)
}

// GetRhsFromStart returns the right hand side from the start.
//
// Parameters:
//   - start: The start.
//
// Returns:
//   - T: The right hand side.
//   - bool: True if the index is valid. False otherwise.
func (item *Item[T]) GetRhsFromStart(start int) (T, bool) {
	uc.Assert(item.rule != nil, "item.rule must not be nil")

	return item.rule.GetRhsFromStart(start)
}
