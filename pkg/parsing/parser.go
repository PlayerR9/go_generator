package parsing

import (
	"errors"
	"fmt"

	uc "github.com/PlayerR9/MyGoLib/Units/common"
	utpx "github.com/PlayerR9/go_generator/util/parsing"
)

// Parser is a parser for the template.
type Parser struct {
	// tokens is the list of tokens.
	tokens []*utpx.Token[TokenType]

	// stack is the parser stack.
	stack *utpx.Stack[TokenType]
}

/*
// check_rule is a helper function that checks the rule.
//
// Parameters:
//   - rule: The rule.
//
// Returns:
//   - error: An error if the rule is invalid.
//
// Errors:
//   - *parsing.ErrExpected: If the rule is invalid.
//
// Assertions:
//   - The rule must not be empty.
func (p *Parser) check_rule(rule []TokenType) error {
	uc.AssertParam("rule", len(rule) > 0, uc.NewErrEmpty(rule))

	var prev *TokenType

	for _, rhs := range rule {
		top, ok := p.stack.Pop()
		if !ok {
			return utpx.NewErrExpected(nil, prev, rhs)
		} else if top.Type != rhs {
			return utpx.NewErrExpected(&top.Type, prev, rhs)
		}

		prev = &top.Type
	}

	return nil
}

// decision is a helper function that decides the next action.
//
// Returns:
//   - utpx.Actioner: The next action.
//   - error: An error if the input stream is invalid.
//
// Assertions:
//   - The stack must not be empty.
func (p *Parser) decision(la *utpx.Token[TokenType]) (utpx.Actioner[TokenType], error) {
	top1, ok := p.stack.Pop()
	uc.Assert(ok, "p.stack.Pop() failed")

	var act utpx.Actioner[TokenType]

	items, ok := AllItems[top1.Type]
	if !ok {
		return nil, fmt.Errorf("unexpected token (%q)", top1.Type.String())
	}

	switch top1.Type {
	case TkEOF:
		// [ EOF ] Source1 -> Source : accept .
		rule := []TokenType{TkEOF, TkSource1}

		err := p.check_rule(rule)
		if err != nil {
			return nil, err
		}

		act = utpx.NewActAccept(TkSource, rule)
	case TkSource1:
		top2, ok := p.stack.Pop()
		if !ok || top2.Type != TkElem {
			// EOF [ Source1 ] -> Source : shift .
			act = utpx.NewActShift[TokenType]()
		} else {
			// [ Source1 ] Elem -> Source1 : reduce .
			act = utpx.NewActReduce(TkSource1, []TokenType{TkSource1, TkElem})
		}
	case TkElem:
		if la == nil {
			// [ Elem ] -> Source1 : reduce .
			rule := []TokenType{TkElem}

			err := p.check_rule(rule)
			if err != nil {
				return nil, err
			}

			act = utpx.NewActReduce(TkSource1, rule)
		} else {
			switch la.Type {
			case TkOpCurly, TkText, TkWs:
				// Source1 [ Elem ] -> Source1 : shift .
				// -- op_curly
				// -- text
				// -- ws
				act = utpx.NewActShift[TokenType]()
			default:
				// [ Elem ] -> Source1 : reduce .
				rule := []TokenType{TkElem}

				err := p.check_rule(rule)
				if err != nil {
					return nil, err
				}

				act = utpx.NewActReduce(TkSource1, []TokenType{TkElem})
			}
		}
	case TkVariable:
		// [ Variable ] -> Elem : reduce .
		rule := []TokenType{TkVariable}

		err := p.check_rule(rule)
		if err != nil {
			return nil, err
		}

		act = utpx.NewActReduce(TkElem, rule)
	case TkText:
		// [ text ] -> Elem : reduce .
		rule := []TokenType{TkText}

		err := p.check_rule(rule)
		if err != nil {
			return nil, err
		}

		act = utpx.NewActReduce(TkElem, rule)
	case TkClCurly:
		top2, ok := p.stack.Pop()
		if !ok {
			return nil, utpx.NewErrExpected(nil, &top1.Type, TkVariableName, TkSws)
		}

		switch top2.Type {
		case TkVariableName:
			top3, ok := p.stack.Pop()
			if !ok {
				return nil, utpx.NewErrExpected(nil, &top2.Type, TkDot)
			} else if top3.Type != TkDot {
				return nil, utpx.NewErrExpected(&top3.Type, &top2.Type, TkDot)
			}

			top4, ok := p.stack.Pop()
			if !ok {
				return nil, utpx.NewErrExpected(nil, &top3.Type, TkOpCurly, TkSws)
			}

			switch top4.Type {
			case TkOpCurly:
				// [ cl_curly ] variable_name dot op_curly -> Variable : reduce .
				act = utpx.NewActReduce(TkVariable, []TokenType{TkClCurly, TkVariableName, TkDot, TkOpCurly})
			case TkSws:
				// [ cl_curly ] variable_name dot Sws op_curly -> Variable : reduce .
				act = utpx.NewActReduce(TkVariable, []TokenType{TkClCurly, TkVariableName, TkDot, TkSws, TkOpCurly})
			default:
				return nil, utpx.NewErrExpected(&top4.Type, &top3.Type, TkOpCurly, TkSws)
			}
		case TkSws:
			top3, ok := p.stack.Pop()
			if !ok {
				return nil, utpx.NewErrExpected(nil, &top2.Type, TkVariableName)
			} else if top3.Type != TkVariableName {
				return nil, utpx.NewErrExpected(&top3.Type, &top2.Type, TkVariableName)
			}

			top4, ok := p.stack.Pop()
			if !ok {
				return nil, utpx.NewErrExpected(nil, &top3.Type, TkDot)
			} else if top4.Type != TkDot {
				return nil, utpx.NewErrExpected(&top4.Type, &top3.Type, TkDot)
			}

			top5, ok := p.stack.Pop()
			if !ok {
				return nil, utpx.NewErrExpected(nil, &top4.Type, TkOpCurly, TkSws)
			}

			switch top5.Type {
			case TkOpCurly:
				// [ cl_curly ] Sws variable_name dot op_curly -> Variable : reduce .
				act = utpx.NewActReduce(TkVariable, []TokenType{TkClCurly, TkSws, TkVariableName, TkDot, TkOpCurly})
			case TkSws:
				// [ cl_curly ] Sws variable_name dot Sws op_curly -> Variable : reduce .
				act = utpx.NewActReduce(TkVariable, []TokenType{TkClCurly, TkSws, TkVariableName, TkDot, TkSws, TkOpCurly})
			default:
				return nil, utpx.NewErrExpected(&top5.Type, &top4.Type, TkOpCurly, TkSws)
			}
		default:
			return nil, utpx.NewErrExpected(&top2.Type, &top1.Type, TkVariableName, TkSws)
		}
	case TkVariableName:
		// cl_curly [ variable_name ] dot op_curly -> Variable : shift .
		// cl_curly [ variable_name ] dot Sws op_curly -> Variable : shift .
		// cl_curly Sws [ variable_name ] dot op_curly -> Variable : shift .
		// cl_curly Sws [ variable_name ] dot Sws op_curly -> Variable : shift .
		act = utpx.NewActShift[TokenType]()
	case TkDot:
		// cl_curly variable_name [ dot ] op_curly -> Variable : shift .
		// cl_curly variable_name [ dot ] Sws op_curly -> Variable : shift .
		// cl_curly Sws variable_name [ dot ] op_curly -> Variable : shift .
		// cl_curly Sws variable_name [ dot ] Sws op_curly -> Variable : shift .
		act = utpx.NewActShift[TokenType]()
	case TkOpCurly:
		// cl_curly variable_name dot [ op_curly ] -> Variable : shift .
		// cl_curly variable_name dot Sws [ op_curly ] -> Variable : shift .
		// cl_curly Sws variable_name dot [ op_curly ] -> Variable : shift .
		// cl_curly Sws variable_name dot Sws [ op_curly ] -> Variable : shift .
		act = utpx.NewActShift[TokenType]()
	case TkSws:
		if la == nil {
			top2, ok := p.stack.Pop()
			if !ok || top2.Type != TkWs {
				// [ Sws ] -> Elem : reduce .
				rule := []TokenType{TkSws}
				act = utpx.NewActReduce(TkElem, []TokenType{TkSws})
			} else {
				// [ Sws ] ws -> Sws : reduce .
				act = utpx.NewActReduce(TkSws, []TokenType{TkSws, TkWs})
			}
		} else {
			switch la.Type {
			case TkDot, TkClCurly:
				// cl_curly variable_name dot [ Sws ] op_curly -> Variable : shift .
				// cl_curly [ Sws ] variable_name dot op_curly  -> Variable : shift .
				// cl_curly Sws variable_name dot [ Sws ] op_curly -> Variable : shift .
				act = utpx.NewActShift[TokenType]()
			default:
				top2, ok := p.stack.Pop()
				if !ok || top2.Type != TkWs {
					// [ Sws ] -> Elem : reduce .
					act = utpx.NewActReduce(TkElem, []TokenType{TkSws})
				} else {
					// [ Sws ] ws -> Sws : reduce .
					act = utpx.NewActReduce(TkSws, []TokenType{TkSws, TkWs})
				}
			}
		}
	case TkWs:
		if la == nil || la.Type != TkWs {
			// [ ws ] -> Sws : reduce .
			act = utpx.NewActReduce(TkSws, []TokenType{TkWs})
		} else {
			// Sws [ ws ] -> Sws : shift .
			// -- ws
			act = utpx.NewActShift[TokenType]()
		}
	default:
		return nil, fmt.Errorf("unexpected token %s", top1.Type)
	}

	return act, nil
}
*/

// shift is a helper function that shifts the input stream.
//
// Returns:
//   - bool: True if the input stream is valid, false otherwise.
func (p *Parser) shift() bool {
	if len(p.tokens) == 0 {
		return false
	}

	first := p.tokens[0]
	p.tokens = p.tokens[1:]

	p.stack.Push(first)

	return true
}

// reduce is a helper function that reduces the input stream with the given action.
//
// Parameters:
//   - act: The action to reduce the input stream with.
//
// Returns:
//   - error: An error if the input stream is invalid.
//
// Assertions:
//   - The action must not be nil.
//   - The iterator of act must not be nil.
func (p *Parser) reduce(act utpx.Actioner[TokenType]) error {
	uc.AssertParam("act", act != nil, errors.New("act is nil"))

	lhs := act.GetLHS()

	iter := act.Iterator()
	uc.Assert(iter != nil, "iterator should not be nil")

	var prev *TokenType

	for {
		curr, err := iter.Consume()
		if err != nil {
			break
		}

		top, ok := p.stack.Pop()
		if !ok {
			p.stack.RefuseMany()

			return utpx.NewErrReduce(lhs, curr, prev, nil)
		} else if top.Type != curr {
			p.stack.RefuseMany()

			return utpx.NewErrReduce(lhs, curr, prev, &top.Type)
		}

		prev = &curr
	}

	popped := p.stack.GetPopped()
	p.stack.Accept()

	tk := utpx.NewToken(lhs, popped, popped[len(popped)-1].Lookahead)
	p.stack.Push(tk)

	return nil
}

// apply_action is a helper function that applies the given action.
//
// Parameters:
//   - act: The action to apply.
//
// Returns:
//   - bool: True if the action is an accept action, false otherwise.
//   - error: An error if the action was not applied successfully.
//
// Assertions:
//   - The action must not be nil.
func (p *Parser) apply_action(act utpx.Actioner[TokenType]) (bool, error) {
	uc.AssertParam("act", act != nil, errors.New("act is nil"))

	switch act := act.(type) {
	case *utpx.ActAccept[TokenType]:
		err := p.reduce(act)
		if err != nil {
			return false, fmt.Errorf("accept failed: %w", err)
		}

		return true, nil
	case *utpx.ActReduce[TokenType]:
		err := p.reduce(act)
		if err != nil {
			if DebugMode {
				fmt.Println("Token tree:")

				for {
					top, ok := p.stack.Pop()
					if !ok {
						break
					}

					fmt.Println(utpx.PrintTokenTree(top))
					fmt.Println()
				}
				fmt.Println()
			}
			return false, fmt.Errorf("reduce failed: %w", err)
		}
	case *utpx.ActShift[TokenType]:
		ok := p.shift()
		if !ok {
			return false, fmt.Errorf("shift failed")
		}
	default:
		return false, fmt.Errorf("unexpected action %T", act)
	}

	return false, nil
}

// Parse is a helper function that parses the given tokens.
//
// Parameters:
//   - tokens: The tokens to parse.
//
// Returns:
//   - *utpx.Token[TokenType]: The parsed token.
//   - error: An error if the parsing failed.
//
// Assertions:
//   - The initial shift must not fail.
//   - At the end of the parsing process, the stack must not be empty.
func Parse(tokens []*utpx.Token[TokenType]) (*utpx.Token[TokenType], error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("no tokens")
	}

	p := &Parser{
		stack:  utpx.NewStack[TokenType](),
		tokens: tokens,
	}

	ok := p.shift() // Initial shift
	uc.Assert(ok, "initial shift failed")

	for {
		top, ok := p.stack.Peek()
		if !ok {
			// FIXME: Check this control flow.
			break
		}

		act, err := DecisionTable.Decide(p.stack, &top.Lookahead.Type)
		p.stack.RefuseMany()
		if err != nil {
			if DebugMode {
				// DEBUG: Print token tree
				fmt.Println("Token tree:")

				for {
					top, ok := p.stack.Pop()
					if !ok {
						break
					}

					fmt.Println(utpx.PrintTokenTree(top))
					fmt.Println()
				}
				fmt.Println()
			}

			return nil, fmt.Errorf("could not decide: %w", err)
		}

		is_done, err := p.apply_action(act)
		if err != nil {
			return nil, fmt.Errorf("could not apply action: %w", err)
		}

		if is_done {
			break
		}
	}

	top, ok := p.stack.Pop()
	uc.Assert(ok, "no top element on the stack")

	if !p.stack.IsEmpty() {
		if DebugMode {
			// DEBUG: Print token tree
			for {
				top, ok := p.stack.Pop()
				if !ok {
					break
				}

				fmt.Println("Token tree:")
				fmt.Println(utpx.PrintTokenTree(top))
				fmt.Println()
			}

			fmt.Println()
		}

		return nil, fmt.Errorf("some elements are left on the stack")
	}

	if len(p.tokens) != 0 {
		Logger.Println("Parsing ended early")
	}

	return top, nil
}
