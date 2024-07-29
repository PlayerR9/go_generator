package parsing

import (
	"strconv"
	"strings"

	utstr "github.com/PlayerR9/lib_units/strings"
)

// ErrExpected is an error for expected values.
type ErrExpected[T TokenTyper] struct {
	// Expecteds is the expected value.
	Expecteds []T

	// Got is the actual value.
	Got *T

	// Before is the value before the error.
	Before *T
}

// Error implements the error interface.
//
// Message: "expected {{ .Expected }} before {{ .Before }}, got {{ .Got }} instead."
func (e *ErrExpected[T]) Error() string {
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

	builder.WriteString("expected ")

	values := make([]string, 0, len(e.Expecteds))

	for _, v := range e.Expecteds {
		values = append(values, strconv.Quote(v.String()))
	}

	builder.WriteString(utstr.OrString(values, false, false))
	builder.WriteString(" before ")
	builder.WriteString(before)
	builder.WriteString(", got ")
	builder.WriteString(got)
	builder.WriteString(" instead")

	return builder.String()
}

// NewErrExpected creates a new error.
//
// Parameters:
//   - expecteds: The expected values.
//   - got: The actual value.
//   - before: The value before the error.
//
// Returns:
//   - *ErrExpected[T]: The error. Never returns nil.
func NewErrExpected[T TokenTyper](got *T, before *T, expecteds ...T) *ErrExpected[T] {
	return &ErrExpected[T]{
		Expecteds: expecteds,
		Got:       got,
		Before:    before,
	}
}

// ErrReduce is an error that occurs during reduction.
type ErrReduce[T TokenTyper] struct {
	// Lhs is the left hand side.
	Lhs T

	// Expected is the expected value.
	Expected T

	// Got is the actual value.
	Got *T

	// Before is the value before the error.
	Before *T
}

// Error implements the error interface.
//
// Message: "after {{ .Lhs }}: expected {{ .Expected }} before {{ .Before }}, got {{ .Got }} instead."
func (e *ErrReduce[T]) Error() string {
	var got string

	if e.Got == nil {
		got = "nothing"
	} else {
		got = strconv.Quote((*e.Got).String())
	}

	var builder strings.Builder

	builder.WriteString("after ")
	builder.WriteString(strconv.Quote(e.Lhs.String()))
	builder.WriteString(": expected ")
	builder.WriteString(strconv.Quote(e.Expected.String()))

	if e.Before != nil {
		builder.WriteString(" before ")
		builder.WriteString(strconv.Quote((*e.Before).String()))
	}

	builder.WriteString(", got ")
	builder.WriteString(got)
	builder.WriteString(" instead")

	return builder.String()
}

// NewErrReduce creates a new error.
//
// Parameters:
//   - lhs: The left hand side.
//   - expected: The expected value.
//   - got: The actual value.
//   - before: The value before the error.
//
// Returns:
//   - *ErrReduce[T]: The error. Never returns nil.
func NewErrReduce[T TokenTyper](lhs T, expected T, got *T, before *T) *ErrReduce[T] {
	return &ErrReduce[T]{
		Lhs:      lhs,
		Expected: expected,
		Got:      got,
		Before:   before,
	}
}
