package parsing

import (
	"strconv"
	"testing"
)

type TestTokenType int

const (
	Elem TestTokenType = iota
)

func (t TestTokenType) IsAcceptSymbol() bool {
	return false
}

func (t TestTokenType) IsTerminal() bool {
	return false
}

func (t TestTokenType) String() string {
	return [...]string{
		"Elem",
	}[t]
}

func TestPeek(t *testing.T) {
	stack := NewStack[TestTokenType]()

	stack.Push(nil)

	_, ok := stack.Peek()
	if ok {
		t.Errorf("expected false, got true")
	}

	tok := NewToken(Elem, "", nil)
	stack.Push(tok)

	tok, ok = stack.Peek()
	if !ok {
		t.Errorf("expected true, got false")
	} else if tok.Type != Elem {
		t.Errorf("expected Elem, got %s", tok.Type)
	}
}

func TestPush(t *testing.T) {
	stack := NewStack[TestTokenType]()

	tok1 := NewToken(Elem, "1", nil)
	tok2 := NewToken(Elem, "2", nil)
	tok3 := NewToken(Elem, "3", nil)

	stack.Push(tok1)
	stack.Push(tok2)
	stack.Push(nil)
	stack.Push(tok3)

	tok, ok := stack.Pop()
	if !ok {
		t.Errorf("expected true, got false")
	} else if tok.Type != Elem || tok.Data != "3" {
		t.Errorf("expected Elem, got %s", tok.Type)
	}

	tok, ok = stack.Pop()
	if !ok {
		t.Errorf("expected true, got false")
	} else if tok.Type != Elem || tok.Data != "2" {
		t.Errorf("expected Elem, got %s", tok.Type)
	}

	tok, ok = stack.Pop()
	if !ok {
		t.Errorf("expected true, got false")
	} else if tok.Type != Elem || tok.Data != "1" {
		t.Errorf("expected Elem, got %s", tok.Type)
	}
}

func TestPop(t *testing.T) {
	stack := NewStack[TestTokenType]()

	_, ok := stack.Pop()
	if ok {
		t.Errorf("expected false, got true")
	}
}

func TestRefuseOne(t *testing.T) {
	stack := NewStack[TestTokenType]()

	tok := NewToken(Elem, "1", nil)
	stack.Push(tok)

	res_tok, ok := stack.Pop()
	if !ok {
		t.Errorf("expected true, got false")
	} else if res_tok.Type != Elem || res_tok.Data != "1" {
		t.Errorf("expected Elem, got %s", res_tok.Type)
	}

	ok = stack.RefuseOne()
	if !ok {
		t.Errorf("could not refuse one")
	}

	res_tok, ok = stack.Pop()
	if !ok {
		t.Errorf("expected true, got false")
	} else if res_tok.Type != Elem || res_tok.Data != "1" {
		t.Errorf("expected Elem, got %s", res_tok.Type)
	}
}

func TestRefuseMany(t *testing.T) {
	stack := NewStack[TestTokenType]()

	for i := 0; i < 10; i++ {
		tok := NewToken(Elem, strconv.Itoa(i), nil)

		stack.Push(tok)
	}

	for i := 0; i < 10; i++ {
		stack.Pop()
	}

	stack.RefuseMany()

	size := stack.Size()

	if size != 10 {
		t.Errorf("expected 10, got %d", size)
	}
}

func TestAccept(t *testing.T) {
	stack := NewStack[TestTokenType]()

	for i := 0; i < 10; i++ {
		tok := NewToken(Elem, strconv.Itoa(i), nil)

		stack.Push(tok)
	}

	_, ok := stack.Pop()
	if !ok {
		t.Errorf("expected true, got false")
	}

	stack.Accept()

	size := stack.Size()

	if size != 9 {
		t.Errorf("expected 9, got %d", size)
	}
}
