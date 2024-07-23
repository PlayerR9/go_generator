package parsing

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	uc "github.com/PlayerR9/MyGoLib/Units/common"
	utmp "github.com/PlayerR9/go_generator/util/maps"
)

// DecisionTable is a decision table.
type DecisionTable[T TokenTyper] struct {
	// symbols is the list of symbols.
	symbols []T

	// rules is the list of rules.
	rules []*Rule[T]

	// table is the list of items.
	table map[T][]*Item[T]
}

// parse_rule is a helper function that creates a new rule.
//
// Parameters:
//   - str: The rule string.
//   - f: The function that transforms a field into a token type.
//
// Returns:
//   - *Rule: The new rule. Nil if an error occurs.
//   - error: An error if the rule is invalid.
//
// Assertions:
//   - f must not be nil.
//   - str must not be empty.
func parse_rule[T TokenTyper](str string, f StringToTypeFunc[T]) (*Rule[T], error) {
	uc.AssertParam("f", f != nil, errors.New("value must not be nil"))
	uc.AssertParam("str", str != "", uc.NewErrEmpty(str))

	fields := strings.Fields(str)

	idx := slices.Index(fields, "=")
	if idx == -1 {
		return nil, errors.New("missing \"right arrow\" symbol")
	}

	left := fields[:idx]
	if len(left) == 0 {
		return nil, errors.New("empty left hand side")
	} else if len(left) > 1 {
		return nil, fmt.Errorf("expected only one left hand side, got %d instead", len(left))
	}

	lhs, ok := f(left[0])
	if !ok {
		return nil, fmt.Errorf("invalid left hand side: %s", left[0])
	}

	right := fields[idx+1:]
	if len(right) == 0 {
		return nil, errors.New("empty right hand side")
	}

	rhss := make([]T, 0, len(right))

	for i, field := range right {
		rhs, ok := f(field)
		if !ok {
			return nil, uc.NewErrAt(i+1, "field", fmt.Errorf("invalid right hand side: %s", field))
		}

		rhss = append(rhss, rhs)
	}

	slices.Reverse(rhss)

	r := NewRule(lhs, rhss)
	uc.Assert(r != nil, "invalid rule")

	return r, nil
}

// parse_rules is a helper function that parses the grammar rules.
//
// Parameters:
//   - str: The grammar string.
//   - f: The function that transforms a field into a token type.
//
// Returns:
//   - []*Rule: The rules.
//   - error: An error if the grammar is invalid.
//
// Assertions:
//   - f is not nil.
//   - str is not empty.
func parse_rules[T TokenTyper](str string, f StringToTypeFunc[T]) ([]*Rule[T], error) {
	uc.AssertParam("f", f != nil, errors.New("value must not be nil"))
	uc.AssertParam("str", str != "", uc.NewErrEmpty(str))

	lines := strings.Split(str, ".\n")

	var rules []*Rule[T]

	for i, line := range lines {
		if line == "" {
			continue
		}

		r, err := parse_rule[T](line, f)
		if err != nil {
			return nil, uc.NewErrWhileAt("parsing", i+1, "line", err)
		}

		rules = append(rules, r)
	}

	return rules, nil
}

// make_symbols is a helper function that returns the symbols in the grammar; ignoring duplicates.
func (dt *DecisionTable[T]) make_symbols() {
	uc.Assert(len(dt.rules) > 0, "rules must not be empty")

	var symbols []T

	for _, rule := range dt.rules {
		for _, rhs := range rule.rhss {
			pos, ok := slices.BinarySearch(symbols, rhs)
			if !ok {
				symbols = slices.Insert(symbols, pos, rhs)
			}
		}
	}

	dt.symbols = symbols
}

// make_items is a helper function that creates all possible items for the given rules and symbols.
func (dt *DecisionTable[T]) make_items() {
	uc.Assert(len(dt.rules) > 0, "rules must not be empty")
	uc.Assert(len(dt.symbols) > 0, "symbols must not be empty")

	item_table := make(map[T][]*Item[T])

	for _, symbol := range dt.symbols {
		var items []*Item[T]

		for _, rule := range dt.rules {
			indices := rule.GetIndicesOfRhs(symbol)
			if len(indices) == 0 {
				continue
			}

			for _, idx := range indices {
				item, err := NewItem(rule, idx, nil)
				uc.AssertErr(err, "NewItem(%q, %d)", rule.String(), idx)

				items = append(items, item)
			}
		}

		item_table[symbol] = items
	}

	dt.table = item_table
}

// NewDecisionTable creates a new decision table.
//
// Parameters:
//   - grammar: The grammar string.
//   - f: The function that transforms a field into a token type.
//
// Returns:
//   - *DecisionTable: The new decision table.
//   - error: An error if the grammar is invalid.
func NewDecisionTable[T TokenTyper](grammar string, f StringToTypeFunc[T]) (*DecisionTable[T], error) {
	if f == nil {
		return nil, uc.NewErrNilParameter("f")
	} else if grammar == "" {
		return nil, uc.NewErrInvalidParameter("grammar", uc.NewErrEmpty(grammar))
	}

	dt := &DecisionTable[T]{}

	rules, err := parse_rules(grammar, f)
	if err != nil {
		return nil, err
	}

	dt.rules = rules
	dt.make_symbols()
	dt.make_items()

	if DebugMode {
		// DEBUG: Print the decision table
		fmt.Println("Decision Table:")

		for _, items := range dt.table {
			for _, item := range items {
				fmt.Println(item.String())
			}
			fmt.Println()
		}
		fmt.Println()
	}

	return dt, nil
}

// Decide is a helper function that decides the next action.
//
// Parameters:
//   - stack: The stack.
//   - la: The lookahead token.
//
// Returns:
//   - Actioner: The next action.
//   - error: An error if the input stream is invalid.
func (dt *DecisionTable[T]) Decide(stack *Stack[T], la *T) (Actioner[T], error) {
	if stack == nil {
		return nil, uc.NewErrNilParameter("stack")
	}

	top1, ok := stack.Pop()
	if !ok {
		return nil, uc.NewErrInvalidParameter("stack", uc.NewErrEmpty(stack))
	}

	items, ok := dt.table[top1.Type]
	if !ok {
		return nil, fmt.Errorf("unexpected token (%q)", top1.Type.String())
	}

	uc.Assert(len(items) > 0, "items must not be empty")

	if len(items) == 1 {
		item := items[0]

		err := item.MatchLookahead(la)
		if err != nil {
			return nil, fmt.Errorf("at rule (%q): %w", item.GetLhs().String(), err)
		}

		return item.action, nil
	}

	var matched_items []*Item[T]

	for _, k := range items {
		err := k.MatchLookahead(la)
		if err == nil {
			matched_items = append(matched_items, k)
		}
	}

	if DebugMode {
		fmt.Println("matched_items:")

		for _, item := range matched_items {
			fmt.Println(item.String())
		}
		fmt.Println()
	}

	var err_sol uc.ErrOrSol[*Item[T]]

	for delta := 0; len(matched_items) > 0; delta++ {
		var top_idx int

		for i := 0; i < len(matched_items); i++ {
			item := matched_items[i]

			rhs, ok := item.GetRhsRelative(delta)
			uc.AssertOk(ok, "k.GetRhsRelative(%d) failed", delta)

			top, ok := stack.Pop()
			if !ok {
				err := fmt.Errorf("expected %q, got nothing instead", rhs.String())
				err_sol.AddErr(err, delta)

				continue
			}

			if top.Type != rhs {
				err := fmt.Errorf("expected %q, got %q instead", rhs.String(), top.Type.String())
				err_sol.AddErr(err, delta)

				continue
			}

			if item.IsDone(delta) {
				err_sol.AddSol(item, delta)

				matched_items[i] = item
				top_idx++
			}
		}

		matched_items = matched_items[:top_idx]
	}

	if err_sol.HasError() {
		errs := err_sol.GetErrors()

		if len(errs) == 1 {
			return nil, errs[0]
		} else {
			// FIXME: Return an error that is the union of all errors.
			// However, as of now, only the first error is returned.
			return nil, errs[0]
		}
	}

	sols := err_sol.GetSolutions()
	if len(sols) == 0 {
		return nil, errors.New("no solution")
	}

	if len(sols) == 1 {
		return sols[0].action, nil
	}

	// FIXME: Return the most likely solution. However,
	// as of now, we return the ambiguous grammar error.

	return nil, errors.New("ambiguous grammar")

	/*
		var act Actioner[T]

		switch top1.Type {
		case TkElem:
			if la == nil {
				// [ Elem ] -> Source1 : reduce .
				rule := []T{TkElem}

				act = NewActReduce(TkSource1, rule)
			} else {
				switch la.Type {
				case TkOpCurly, TkText, TkWs:
					// Source1 [ Elem ] -> Source1 : shift .
					// -- op_curly
					// -- text
					// -- ws
					act = NewActShift[T]()
				default:
					// [ Elem ] -> Source1 : reduce .
					rule := []T{TkElem}

					err := p.check_rule(rule)
					if err != nil {
						return nil, err
					}

					act = NewActReduce(TkSource1, []T{TkElem})
				}
			}
		default:
			return nil, fmt.Errorf("unexpected token %s", top1.Type)
		}
	*/
}

// GetItemsByLhs returns all the items whose LHS is 'lhs'.
//
// Parameters:
//   - lhs: The lhs to search for.
//
// Returns:
//   - []*Item: All the items whose LHS is 'lhs'.
func (dt *DecisionTable[T]) GetItemsByLhs(lhs T) []*Item[T] {
	var sol []*Item[T]

	for _, items := range dt.table {
		for _, item := range items {
			if item.GetLhs() == lhs {
				sol = append(sol, item)
			}
		}
	}

	return sol
}

func (dt *DecisionTable[T]) solve_conflicts() {
	target_items := make(map[*Item[T]]T)

	for _, items := range dt.table {
		if len(items) == 0 {
			continue
		}

		for _, item := range items {
			prev, ok := item.GetRhsRelative(-1)
			if ok && !prev.IsTerminal() {
				target_items[item] = prev
			}
		}
	}

	for item, prev := range target_items {
		seen := utmp.NewSeenMap[*Item[T]]()

		la := dt.solve_lookaheads(seen, prev)
	}
}

func (dt *DecisionTable[T]) solve_lookaheads(seen *utmp.SeenMap[*Item[T]], target T) []T {
	uc.AssertParam("seen", seen != nil, errors.New("value must not be nil"))

	// 1. Find all rules whose LHS is the same as the 'target'.
	other_items := dt.GetItemsByLhs(target)
	uc.AssertF(len(other_items) > 0, "no rule has LHS of %q", target.String())

	// 2. Ensure that a role does not call itself. (This prevents infinite loops.)
	other_items = seen.FilterSeen(other_items)
	if len(other_items) == 0 {
		return nil
	}

	var symbols []T
	var to_seek []T

	for _, other_item := range other_items {
		start_rhs, ok := other_item.GetRhsFromStart(0)
		uc.AssertOk(ok, "other_item.GetRhsFromStart(0) failed")

		if start_rhs.IsTerminal() {
			pos, ok := slices.BinarySearch(symbols, start_rhs)
			if !ok {
				symbols = slices.Insert(symbols, pos, start_rhs)
			}
		} else {
			pos, ok := slices.BinarySearch(to_seek, start_rhs)
			if !ok {
				to_seek = slices.Insert(to_seek, pos, start_rhs)
			}
		}
	}

	if len(to_seek) == 0 {
		return symbols
	}

	for _, seek := range to_seek {
		other_symbols := dt.solve_lookaheads(seen, seek)

		for _, symbol := range other_symbols {
			pos, ok := slices.BinarySearch(symbols, symbol)
			if !ok {
				symbols = slices.Insert(symbols, pos, symbol)
			}
		}
	}

	return symbols
}
