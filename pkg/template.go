package pkg

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	uc "github.com/PlayerR9/MyGoLib/Units/common"
	prx "github.com/PlayerR9/go_generator/pkg/parsing"
	utpx "github.com/PlayerR9/go_generator/util/parsing"
)

// Template is a template.
type Template struct {
	// root is the root node of the AST.
	root *Node
}

// NewTemplate creates a new template.
//
// Parameters:
//   - str: The template string.
//
// Returns:
//   - *Template: The template. Nil if an error occurs.
//   - error: An error if the template is invalid.
func NewTemplate(str string) (*Template, error) {
	tokens, err := prx.Lex(str)
	if err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}

	if DebugMode {
		// DEBUG: Print the tokens
		fmt.Println("Tokens:")

		for _, token := range tokens {
			fmt.Println(token.GoString())
		}

		fmt.Println()
	}

	root, err := prx.Parse(tokens)
	if err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}

	if DebugMode {
		// DEBUG: Print the parsed tree
		fmt.Println("Tree:")
		fmt.Println(utpx.PrintTokenTree(root))
		fmt.Println()
	}

	node, err := ToAST(root)
	if err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}

	if DebugMode {
		// DEBUG: Print the AST
		fmt.Println("AST:")
		fmt.Println(PrintAST(node))
		fmt.Println()
	}

	return &Template{
		root: node,
	}, nil
}

func (t *Template) Apply(data any) error {
	if data == nil {
		return uc.NewErrNilParameter("data")
	}

	value := reflect.ValueOf(data)

	for !value.IsValid() {
		kind := value.Kind()

		if kind != reflect.Interface && kind != reflect.Pointer {
			break
		}

		value = value.Elem()
	}

	if !value.IsValid() {
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
	uc.AssertParam("value", value.IsValid(), errors.New("value is zero"))

	switch node.Kind {
	case VariableNode:
		field := value.FieldByName(node.Data)
		if field.IsValid() {
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
