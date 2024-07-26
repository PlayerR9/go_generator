package parsing

import (
	"fmt"
	"strconv"
	"strings"

	fstr "github.com/PlayerR9/MyGoLib/Formatting/Strings"
	uc "github.com/PlayerR9/MyGoLib/Units/common"
)

type TokenTyper interface {
	~int

	// IsAcceptSymbol returns true if the token is an accept symbol.
	//
	// Returns:
	//   - bool: True if the token is an accept symbol. False otherwise.
	IsAcceptSymbol() bool

	// IsTerminal returns true if the token is a terminal.
	//
	// Returns:
	//   - bool: True if the token is a terminal. False otherwise.
	IsTerminal() bool

	fmt.GoStringer
	fmt.Stringer
}

type Token[T TokenTyper] struct {
	Type      T
	Data      any // either string or []*Token[T]
	Lookahead *Token[T]
}

func (t *Token[T]) GoString() string {
	var builder strings.Builder

	builder.WriteString("Token[Type=")
	builder.WriteString(t.Type.GoString())
	builder.WriteString(", Data=")

	switch data := t.Data.(type) {
	case string:
		builder.WriteString(strconv.Quote(data))
	case []*Token[T]:
		values := make([]string, 0, len(data))

		for _, tok := range data {
			values = append(values, fmt.Sprintf("%p", tok))
		}

		builder.WriteRune('[')
		builder.WriteString(strings.Join(values, ", "))
		builder.WriteRune(']')
	default:
		builder.WriteString("unknown")
	}

	builder.WriteString(", Lookahead=")

	if t.Lookahead == nil {
		builder.WriteString("nil")
	} else {
		fmt.Fprintf(&builder, "%p", t.Lookahead)
	}

	builder.WriteString("]")

	return builder.String()
}

func (t *Token[T]) String() string {
	var builder strings.Builder

	switch data := t.Data.(type) {
	case string:
		builder.WriteString("Token[")
		builder.WriteString(t.Type.GoString())
		builder.WriteString("(")
		builder.WriteString(strconv.Quote(data))
		builder.WriteString(")]")
	case []*Token[T]:
		builder.WriteString("Token[")
		builder.WriteString(t.Type.GoString())
		builder.WriteString("]")
	default:
		builder.WriteString("unknown")
	}

	return builder.String()
}

// nil if either the data is nil or it is not a string nor a []*Token[T]
func NewToken[T TokenTyper](typ T, data any, la *Token[T]) *Token[T] {
	if data == nil {
		return nil
	}

	switch data := data.(type) {
	case string, []*Token[T]:
		return &Token[T]{
			Type:      typ,
			Data:      data,
			Lookahead: la,
		}
	default:
		return nil
	}
}

func (t *Token[T]) IsLeaf() bool {
	_, ok := t.Data.(string)
	return ok
}

func (t *Token[T]) Iterator() uc.Iterater[fstr.Noder] {
	children, ok := t.Data.([]*Token[T])
	if !ok {
		return nil
	}

	elems := make([]fstr.Noder, 0, len(children))

	for _, child := range children {
		elems = append(elems, child)
	}

	return uc.NewSimpleIterator(elems)
}
