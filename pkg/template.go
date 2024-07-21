package pkg

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	uc "github.com/PlayerR9/MyGoLib/Units/common"
	prx "github.com/PlayerR9/go_generator/pkg/parsing"
)

type Template struct {
	root *Node
}

func NewTemplate(str string) (*Template, error) {
	tokens, err := prx.Lex(str)
	if err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}

	root, err := prx.Parse(tokens)
	if err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}

	// DEBUG: Print the parsed tree
	// fmt.Println("Tree:")
	// fmt.Println(utpx.PrintTokenTree(root))
	// fmt.Println()

	node, err := ToAST(root)
	if err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}

	// DEBUG: Print the AST
	// fmt.Println("AST:")
	// fmt.Println(PrintAST(node))
	// fmt.Println()

	return &Template{
		root: node,
	}, nil
}

func (t *Template) Apply(data any) error {
	if data == nil {
		return uc.NewErrNilParameter("data")
	}

	value := reflect.ValueOf(data)

	for !value.IsZero() {
		kind := value.Kind()

		if kind != reflect.Interface && kind != reflect.Pointer {
			break
		}

		value = value.Elem()
	}

	if value.IsZero() {
		return fmt.Errorf("invalid data type: %s", value.Type().String())
	}

	for _, node := range t.root.Children {
		err := t.traverse(node, value)
		if err != nil {
			return err
		}
	}

	// Node[Source]
	//  ├── Node[Variable (A)]
	//  ├── Node[Variable (B)]
	//  └── Node[Text (my_type)]

	return nil
}

func (t *Template) traverse(node *Node, value reflect.Value) error {
	uc.AssertParam("node", node != nil, errors.New("node is nil"))
	uc.AssertParam("value", !value.IsZero(), errors.New("value is zero"))

	switch node.Kind {
	case VariableNode:
		field := value.FieldByName(node.Data)
		if !field.IsZero() {
			node.Kind = TextNode
			node.Data = field.String()
		}
	case TextNode:
		// Do nothing
	default:
		return fmt.Errorf("invalid node: %s", node.Kind.String())
	}

	return nil
}

func (t *Template) Write(w io.Writer) error {
	if w == nil {
		return uc.NewErrNilParameter("w")
	}

	for _, node := range t.root.Children {
		switch node.Kind {
		case VariableNode:
			var builder strings.Builder

			builder.WriteString("{{ .")
			builder.WriteString(node.Data)
			builder.WriteString(" }}")

			bytes := []byte(builder.String())

			n, err := w.Write(bytes)
			if err != nil {
				return err
			} else if n != len(bytes) {
				return errors.New("failed to write all bytes")
			}
		case TextNode:
			bytes := []byte(node.Data)

			n, err := w.Write(bytes)
			if err != nil {
				return err
			} else if n != len(bytes) {
				return errors.New("failed to write all bytes")
			}
		default:
			return fmt.Errorf("invalid node: %s", node.Kind.String())
		}
	}

	return nil
}

func (t *Template) Execute(w io.Writer, data any) error {
	err := t.Apply(data)
	if err != nil {
		return fmt.Errorf("failed to apply template: %w", err)
	}

	err = t.Write(w)
	if err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	return nil
}
