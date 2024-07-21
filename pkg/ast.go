package pkg

import (
	"errors"
	"fmt"
	"strings"

	uc "github.com/PlayerR9/MyGoLib/Units/common"
	us "github.com/PlayerR9/MyGoLib/Units/slice"
	prx "github.com/PlayerR9/go_generator/pkg/parsing"
	utpx "github.com/PlayerR9/go_generator/util/parsing"
	uttr "github.com/PlayerR9/go_generator/util/tree"
)

type TokenType int

const (
	SourceNode TokenType = iota
	VariableNode
	TextNode
)

func (t TokenType) String() string {
	return [...]string{
		"Source",
		"Variable",
		"Text",
	}[t]
}

type Node struct {
	Parent *Node

	Kind     TokenType
	Data     string
	Children []*Node
}

func (n *Node) String() string {
	var builder strings.Builder

	builder.WriteString("Node[")
	builder.WriteString(n.Kind.String())

	if n.Data != "" {
		builder.WriteString(" (")
		builder.WriteString(n.Data)
		builder.WriteString(")")
	}

	builder.WriteString("]")

	return builder.String()
}

func (n *Node) IsLeaf() bool {
	return len(n.Children) == 0
}

func (n *Node) Iterator() uc.Iterater[uttr.Noder] {
	if len(n.Children) == 0 {
		return nil
	}

	children := make([]uttr.Noder, 0, len(n.Children))
	for _, c := range n.Children {
		children = append(children, c)
	}

	return uc.NewSimpleIterator(children)
}

func NewNode(kind TokenType, data string) *Node {
	return &Node{
		Kind:     kind,
		Data:     data,
		Children: make([]*Node, 0),
	}
}

func (n *Node) SetChildren(children []*Node) {
	children = us.FilterNilValues(children)
	if len(children) == 0 {
		return
	}

	for i := 0; i < len(children); i++ {
		children[i].Parent = n
	}

	n.Children = children
}

func ToAST(root *utpx.Token[prx.TokenType]) (*Node, error) {
	if root == nil {
		return nil, uc.NewErrNilParameter("root")
	}

	if root.Type != prx.TkSource {
		return nil, fmt.Errorf("expected %q to be a Source node, got %q instead", root.GoString(), prx.TkSource.String())
	}

	children, ok := root.Data.([]*utpx.Token[prx.TokenType])
	if !ok {
		return nil, fmt.Errorf("expected %q to be a non-leaf node, got a leaf node instead", root.GoString())
	} else if len(children) != 2 {
		return nil, fmt.Errorf("expected %q to have 2 children, got %d instead", root.GoString(), len(children))
	}

	nodes, err := to_ast(children[0])
	if err != nil {
		return nil, fmt.Errorf("failed to convert %q: %w", children[0].GoString(), err)
	}

	n := NewNode(SourceNode, "")
	n.SetChildren(nodes)

	// Node[SourceNode]
	//  └── Node[VariableNode("A")]
	//  └── Node[VariableNode("B")]
	//  └── Node[TextNode("my_type")]

	return n, nil
}

func to_ast(root *utpx.Token[prx.TokenType]) ([]*Node, error) {
	uc.AssertParam("root", root != nil, errors.New("root must not be nil"))

	var nodes []*Node

	switch root.Type {
	case prx.TkVariable:

		children, ok := root.Data.([]*utpx.Token[prx.TokenType])
		if !ok {
			return nil, fmt.Errorf("expected %q to be a non-leaf node, got a leaf node instead", root.GoString())
		} else if len(children) != 4 {
			return nil, fmt.Errorf("expected %q to have 4 children, got %d instead", root.GoString(), len(children))
		}

		data, ok := children[2].Data.(string)
		if !ok {
			return nil, fmt.Errorf("expected %q to have a variable name, got %q instead", root.GoString(), children[2].GoString())
		}

		nodes = append(nodes, NewNode(VariableNode, data))
	case prx.TkElem:
		children, ok := root.Data.([]*utpx.Token[prx.TokenType])
		if !ok {
			return nil, fmt.Errorf("expected %q to be a non-leaf node, got a leaf node instead", root.GoString())
		} else if len(children) != 1 {
			return nil, fmt.Errorf("expected %q to have 1 child, got %d instead", root.GoString(), len(children))
		}

		sub_nodes, err := to_ast(children[0])
		if err != nil {
			return nil, fmt.Errorf("failed to convert %q: %w", children[0].GoString(), err)
		}

		nodes = append(nodes, sub_nodes...)
	case prx.TkText:
		data, ok := root.Data.(string)
		if !ok {
			return nil, fmt.Errorf("expected %q to be a leaf node, got a non-leaf node instead", root.GoString())
		}

		nodes = append(nodes, NewNode(TextNode, data))
	case prx.TkSource1:
		sub_nodes, err := lhs_ast(prx.TkSource1, root)
		if err != nil {
			return nil, fmt.Errorf("failed to convert %q: %w", root.GoString(), err)
		} else if len(sub_nodes) == 0 {
			return nil, nil
		}

		nodes = append(nodes, sub_nodes...)
	default:
		return nil, fmt.Errorf("expected %q, got %q instead", prx.TkVariable.String(), root.Type.String())
	}

	return nodes, nil
}

func lhs_ast(lhs prx.TokenType, root *utpx.Token[prx.TokenType]) ([]*Node, error) {
	uc.AssertParam("root", root != nil, errors.New("root must not be nil"))

	var nodes []*Node

	for root != nil {
		if root.Type != lhs {
			return nil, fmt.Errorf("expected %q, got %q instead", lhs.String(), root.Type.String())
		}

		children, ok := root.Data.([]*utpx.Token[prx.TokenType])
		if !ok {
			return nil, fmt.Errorf("expected %q to be a non-leaf node, got a leaf node instead", root.GoString())
		} else if len(children) == 0 || len(children) > 2 {
			return nil, fmt.Errorf("expected %q to have at least 1 and at most 2 children, got %d instead", root.GoString(), len(children))
		}

		sub_nodes, err := to_ast(children[0])
		if err != nil {
			return nil, fmt.Errorf("failed to convert %q: %w", children[0].GoString(), err)
		}

		if len(sub_nodes) > 0 {
			nodes = append(nodes, sub_nodes...)
		}

		children = children[1:]

		if len(children) == 0 {
			root = nil
		} else {
			root = children[0]
		}
	}

	return nodes, nil
}

func PrintAST(node *Node) string {
	if node == nil {
		return ""
	}

	str, err := uttr.PrintTree(node)
	uc.AssertErr(err, "tree.PrintTree(%s)", node.String())

	return str
}
