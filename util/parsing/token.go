package parsing

import (
	"fmt"
	"strconv"
	"strings"

	uc "github.com/PlayerR9/MyGoLib/Units/common"
	uttr "github.com/PlayerR9/go_generator/util/tree"
)

type Token[T uc.Enumer] struct {
	Type      T
	Data      any // either string or []*Token[T]
	Lookahead *Token[T]
}

func (t *Token[T]) GoString() string {
	var builder strings.Builder

	builder.WriteString("Token[Type=")
	builder.WriteString(t.Type.String())
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
		builder.WriteString(t.Type.String())
		builder.WriteString("(")
		builder.WriteString(strconv.Quote(data))
		builder.WriteString(")]")
	case []*Token[T]:
		builder.WriteString("Token[")
		builder.WriteString(t.Type.String())
		builder.WriteString("]")
	default:
		builder.WriteString("unknown")
	}

	return builder.String()
}

// nil if either the data is nil or it is not a string nor a []*Token[T]
func NewToken[T uc.Enumer](typ T, data any, la *Token[T]) *Token[T] {
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

func (t *Token[T]) Iterator() uc.Iterater[uttr.Noder] {
	children, ok := t.Data.([]*Token[T])
	if !ok {
		return nil
	}

	elems := make([]uttr.Noder, 0, len(children))

	for _, child := range children {
		elems = append(elems, child)
	}

	return uc.NewSimpleIterator(elems)
}
