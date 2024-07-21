package parsing

import (
	"strconv"
	"strings"

	uc "github.com/PlayerR9/MyGoLib/Units/common"
)

type ErrReduce[T uc.Enumer] struct {
	Lhs      T
	Expected T
	Got      *T
	Before   *T
}

func (e *ErrReduce[T]) Error() string {
	var got string

	if e.Got == nil {
		got = "nothing"
	} else {
		got = strconv.Quote((*e.Got).String())
	}

	var before string

	if e.Before == nil {
		before = "end of file"
	} else {
		before = strconv.Quote((*e.Before).String())
	}

	var builder strings.Builder

	builder.WriteString("after ")
	builder.WriteString(strconv.Quote(e.Lhs.String()))
	builder.WriteString(": expected ")
	builder.WriteString(strconv.Quote(e.Expected.String()))
	builder.WriteString(" before ")
	builder.WriteString(before)
	builder.WriteString(", got ")
	builder.WriteString(got)
	builder.WriteString(" instead")

	return builder.String()
}

func NewErrReduce[T uc.Enumer](lhs T, expected T, got *T, before *T) *ErrReduce[T] {
	return &ErrReduce[T]{
		Lhs:      lhs,
		Expected: expected,
		Got:      got,
		Before:   before,
	}
}
