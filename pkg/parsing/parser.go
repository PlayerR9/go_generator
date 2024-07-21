package parsing

import (
	"errors"
	"fmt"

	uc "github.com/PlayerR9/MyGoLib/Units/common"
	utpx "github.com/PlayerR9/go_generator/util/parsing"
)

// Source = Elem { Elem } EOF .
//
// Elem = Variable | text .
// Variable = op_curly dot variable_name cl_curly .

// Source = Source1 EOF .
// Source1 = Elem .
// Source1 = Elem Source1 .
//
// Elem = Variable .
// Elem = text .
// Variable = op_curly dot variable_name cl_curly .

// EOF Source1 -> Source .
// Elem -> Source1 .
// Source1 Elem -> Source1 .
//
// Variable -> Elem .
// text -> Elem .
// cl_curly variable_name dot op_curly -> Variable .

type Parser struct {
	tokens []*utpx.Token[TokenType]
	stack  *utpx.Stack[TokenType]
}

func (p *Parser) decision() (utpx.Actioner[TokenType], error) {
	top1, ok := p.stack.Pop()
	uc.Assert(ok, "p.stack.Pop() failed")

	var act utpx.Actioner[TokenType]

	switch top1.Type {
	case TkEOF:
		// [ EOF ] Source1 -> Source : accept .
		act = utpx.NewActAccept(TkSource, []TokenType{TkEOF, TkSource1})
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
		la := top1.Lookahead
		if la == nil {
			// [ Elem ] -> Source1 : reduce .
			act = utpx.NewActReduce(TkSource1, []TokenType{TkElem})
		} else {
			switch la.Type {
			case TkOpCurly, TkText:
				// Source1 [ Elem ] -> Source1 : shift .
				// -- op_curly
				// -- text
				act = utpx.NewActShift[TokenType]()
			default:
				// [ Elem ] -> Source1 : reduce .
				act = utpx.NewActReduce(TkSource1, []TokenType{TkElem})
			}
		}
	case TkVariable:
		// [ Variable ] -> Elem : reduce .
		act = utpx.NewActReduce(TkElem, []TokenType{TkVariable})
	case TkText:
		// [ text ] -> Elem : reduce .
		act = utpx.NewActReduce(TkElem, []TokenType{TkText})
	case TkClCurly:
		// [ cl_curly ] variable_name dot op_curly -> Variable : reduce .
		act = utpx.NewActReduce(TkVariable, []TokenType{TkClCurly, TkVariableName, TkDot, TkOpCurly})
	case TkVariableName:
		// cl_curly [ variable_name ] dot op_curly -> Variable : shift .
		act = utpx.NewActShift[TokenType]()
	case TkDot:
		// cl_curly variable_name [ dot ] op_curly -> Variable : shift .
		act = utpx.NewActShift[TokenType]()
	case TkOpCurly:
		// cl_curly variable_name dot [ op_curly ] -> Variable : shift .
		act = utpx.NewActShift[TokenType]()
	default:
		return nil, fmt.Errorf("unexpected token %s", top1.Type)
	}

	return act, nil
}

func (p *Parser) shift() bool {
	if len(p.tokens) == 0 {
		return false
	}

	first := p.tokens[0]
	p.tokens = p.tokens[1:]

	p.stack.Push(first)

	return true
}

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
		act, err := p.decision()
		p.stack.RefuseMany()
		if err != nil {
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
		// DEGUB: Print token tree
		str := utpx.PrintTokenTree(top)

		fmt.Println(str)

		return nil, fmt.Errorf("some elements are left on the stack")
	}

	if len(p.tokens) != 0 {
		fmt.Println("parsing ended early")
	}

	return top, nil
}
