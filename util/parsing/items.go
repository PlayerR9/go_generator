package parsing

import (
	"fmt"
	"strings"

	uc "github.com/PlayerR9/MyGoLib/Units/common"
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

// Item is an item in the rule.
type Item[T TokenTyper] struct {
	// rule is the rule.
	rule *Rule[T]

	// action is the action.
	action Actioner[T]

	// pos is the position in the rule.
	pos int
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

	return &Item[T]{
		rule:   rule,
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
//   - error: An error if the lookahead does not match the rule.
//
// As a special case, if item does not need lookahead, nil is returned regardless of the lookahead.
func (item *Item[T]) MatchLookahead(la *T) error {
	if item.pos == 0 {
		return nil
	}

	prev, ok := item.rule.GetRhsAt(item.pos - 1)
	uc.AssertOk(ok, "item.rule.GetRhsAt(%d)", item.pos-1)

	if la == nil {
		return fmt.Errorf("expected %q, got nothing instead", prev.String())
	} else if *la != prev {
		return fmt.Errorf("expected %q, got %q instead", prev.String(), *la)
	}

	return nil
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
