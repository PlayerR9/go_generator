package tree

import (
	"fmt"
	"slices"
	"strings"

	uc "github.com/PlayerR9/MyGoLib/Units/common"
)

type Noder interface {
	IsLeaf() bool

	uc.Iterable[Noder]
	fmt.Stringer
}

type StackElement[T Noder] struct {
	indent     string
	node       T
	same_level bool
	is_last    bool
}

type Printer[T Noder] struct {
	lines []string
}

func PrintTree[T Noder](root T) (string, error) {
	p := &Printer[T]{
		lines: make([]string, 0),
	}

	se := &StackElement[T]{
		indent:     "",
		node:       root,
		same_level: false,
		is_last:    true,
	}

	stack := []*StackElement[T]{se}

	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		sub, err := p.trav(top)
		if err != nil {
			return "", err
		} else if len(sub) == 0 {
			continue
		}

		slices.Reverse(sub)

		stack = append(stack, sub...)
	}

	return strings.Join(p.lines, "\n"), nil
}

func (p *Printer[T]) trav(elem *StackElement[T]) ([]*StackElement[T], error) {
	var builder strings.Builder

	if elem.indent != "" {
		builder.WriteString(elem.indent)

		ok := elem.node.IsLeaf()
		if !ok || elem.is_last {
			builder.WriteString("└── ")
		} else {
			builder.WriteString("├── ")
		}
	}

	builder.WriteString(elem.node.String())

	p.lines = append(p.lines, builder.String())

	iter := elem.node.Iterator()
	if iter == nil {
		return nil, nil
	}

	var elems []*StackElement[T]

	var indent strings.Builder

	indent.WriteString(elem.indent)

	if elem.same_level && !elem.is_last {
		indent.WriteString("│   ")
	} else {
		indent.WriteString("    ")
	}

	for {
		value, err := iter.Consume()
		ok := uc.IsDone(err)
		if ok {
			break
		} else if err != nil {
			return nil, err
		}

		node, ok := value.(T)
		if !ok {
			return nil, fmt.Errorf("expected %T, got %T", *new(T), value)
		}

		elems = append(elems, &StackElement[T]{
			indent:     indent.String(),
			node:       node,
			same_level: false,
			is_last:    false,
		})
	}

	if len(elems) == 0 {
		return nil, nil
	}

	if len(elems) >= 2 {
		for i := 0; i < len(elems); i++ {
			elems[i].same_level = true
		}
	}

	elems[len(elems)-1].is_last = true

	return elems, nil
}
