package pkg

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	fstr "github.com/PlayerR9/MyGoLib/Formatting/Strings"
	prx "github.com/PlayerR9/go_generator/pkg/parsing"
	utpx "github.com/PlayerR9/go_generator/util/parsing"
	uc "github.com/PlayerR9/lib_units/common"
	us "github.com/PlayerR9/lib_units/slices"
)

// NodeType is the type of a token.
type NodeType int

const (
	// SourceNode is the source node.
	SourceNode NodeType = iota

	// VariableNode is the variable node.
	VariableNode

	// TextNode is the text node.
	TextNode
)

// String implements the common.Enumer interface.
func (t NodeType) String() string {
	return [...]string{
		"Source",
		"Variable",
		"Text",
	}[t]
}

// Node is the AST node.
type Node struct {
	// Parent is the parent node.
	Parent *Node

	// Kind is the type of the node.
	Kind NodeType

	// Data is the data of the node.
	Data string

	// Children is the list of children nodes.
	Children []*Node
}

// String implements the tree.Noder interface.
//
// Format:
//
//	Node[Kind (Data)]
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

// IsLeaf implements the tree.Noder interface.
func (n *Node) IsLeaf() bool {
	return len(n.Children) == 0
}

// Iterator implements the tree.Noder interface.
//
// Never returns nil
func (n *Node) Iterator() uc.Iterater[fstr.Noder] {
	var children []fstr.Noder

	if len(n.Children) != 0 {
		children = make([]fstr.Noder, 0, len(n.Children))
		for _, c := range n.Children {
			children = append(children, c)
		}
	}

	return uc.NewSimpleIterator(children)
}

// NewNode creates a new node.
//
// Parameters:
//   - kind: The type of the node.
//   - data: The data of the node.
//
// Returns:
//   - *Node: The node. Never returns nil.
func NewNode(kind NodeType, data string) *Node {
	return &Node{
		Kind:     kind,
		Data:     data,
		Children: make([]*Node, 0),
	}
}

// SetChildren sets the children of the node. It skips nil children.
//
// Parameters:
//   - children: The children of the node.
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

// ToAST converts the token tree to an AST.
//
// Parameters:
//   - root: The root token of the tree.
//
// Returns:
//   - *Node: The AST. Never returns nil.
//   - error: An error if the tree is invalid.
func ToAST(root *utpx.Token[prx.TokenType]) (*Node, error) {
	if root == nil {
		return nil, uc.NewErrNilParameter("root")
	}

	if root.Type != prx.TkSource {
		return nil, fmt.Errorf("expected %q to be a Source node, got %q instead", root.String(), prx.TkSource.String())
	}

	children, ok := root.Data.([]*utpx.Token[prx.TokenType])
	if !ok {
		return nil, fmt.Errorf("expected %q to be a non-leaf node, got a leaf node instead", root.String())
	} else if len(children) != 2 {
		return nil, fmt.Errorf("expected %q to have 2 children, got %d instead", root.String(), len(children))
	}

	nodes, err := to_ast(children[0])
	if err != nil {
		return nil, fmt.Errorf("failed to convert %q: %w", children[0].String(), err)
	}

	n := NewNode(SourceNode, "")
	n.SetChildren(nodes)

	for {
		ok := simplify_ast(n)
		if !ok {
			break
		}
	}

	return n, nil
}

// to_ast is a helper function to convert a token tree to an AST.
//
// Parameters:
//   - root: The root token of the tree.
//
// Returns:
//   - []*Node: The AST. Never returns nil.
//   - error: An error if the tree is invalid.
func to_ast(root *utpx.Token[prx.TokenType]) ([]*Node, error) {
	uc.AssertParam("root", root != nil, errors.New("root must not be nil"))

	var nodes []*Node

	switch root.Type {
	case prx.TkVariable:

		children, ok := root.Data.([]*utpx.Token[prx.TokenType])
		if !ok {
			return nil, fmt.Errorf("expected %q to be a non-leaf node, got a leaf node instead", root.String())
		} else if len(children) < 4 || len(children) > 6 {
			return nil, fmt.Errorf("expected %q to have 4-6 children, got %d instead", root.String(), len(children))
		}

		idx := -1

		for i := 0; i < len(children); i++ {
			if children[i].Type == prx.TkVariableName {
				idx = i
				break
			}
		}

		if idx == -1 {
			return nil, fmt.Errorf("expected %q to have a variable name", root.String())
		}

		data, ok := children[idx].Data.(string)
		if !ok {
			return nil, fmt.Errorf("expected %q to have a variable name, got %q instead", root.String(), children[2].String())
		}

		nodes = append(nodes, NewNode(VariableNode, data))
	case prx.TkElem:
		children, ok := root.Data.([]*utpx.Token[prx.TokenType])
		if !ok {
			return nil, fmt.Errorf("expected %q to be a non-leaf node, got a leaf node instead", root.String())
		} else if len(children) != 1 {
			return nil, fmt.Errorf("expected %q to have 1 child, got %d instead", root.String(), len(children))
		}

		sub_nodes, err := to_ast(children[0])
		if err != nil {
			return nil, fmt.Errorf("failed to convert %q: %w", children[0].String(), err)
		}

		nodes = append(nodes, sub_nodes...)
	case prx.TkText:
		data, ok := root.Data.(string)
		if !ok {
			return nil, fmt.Errorf("expected %q to be a leaf node, got a non-leaf node instead", root.String())
		}

		nodes = append(nodes, NewNode(TextNode, data))
	case prx.TkSource1:
		sub_nodes, err := lhs_ast(prx.TkSource1, root)
		if err != nil {
			return nil, fmt.Errorf("failed to convert %q: %w", root.String(), err)
		} else if len(sub_nodes) == 0 {
			return nil, nil
		}

		nodes = append(nodes, sub_nodes...)
	case prx.TkWs:
		data, ok := root.Data.(string)
		if !ok {
			return nil, fmt.Errorf("expected %q to be a leaf node, got a non-leaf node instead", root.String())
		}

		nodes = append(nodes, NewNode(TextNode, data))
	default:
		return nil, utpx.NewErrExpected(&root.Type, nil, prx.TkVariable, prx.TkElem, prx.TkText, prx.TkSource1)
	}

	return nodes, nil
}

// lhs_ast is a helper function to convert a LHS token tree to an AST.
//
// Parameters:
//   - lhs: The LHS token of the tree.
//   - root: The root token of the tree.
//
// Returns:
//   - []*Node: The AST. Never returns nil.
//   - error: An error if the tree is invalid.
func lhs_ast(lhs prx.TokenType, root *utpx.Token[prx.TokenType]) ([]*Node, error) {
	uc.AssertParam("root", root != nil, errors.New("root must not be nil"))

	var nodes []*Node

	for root != nil {
		if root.Type != lhs {
			return nil, fmt.Errorf("expected %q, got %q instead", lhs.String(), root.Type.String())
		}

		children, ok := root.Data.([]*utpx.Token[prx.TokenType])
		if !ok {
			return nil, fmt.Errorf("expected %q to be a non-leaf node, got a leaf node instead", root.String())
		} else if len(children) == 0 || len(children) > 2 {
			return nil, fmt.Errorf("expected %q to have at least 1 and at most 2 children, got %d instead", root.String(), len(children))
		}

		sub_nodes, err := to_ast(children[0])
		if err != nil {
			return nil, fmt.Errorf("failed to convert %q: %w", children[0].String(), err)
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

// simplify_ast is a helper function to simplify the AST.
//
// Parameters:
//   - root: The root node of the tree.
//
// Returns:
//   - bool: True if the tree is simplified, false otherwise.
//
// Assertions:
//   - The root node must not be nil.
//   - No node in the tree can be nil.
func simplify_ast(root *Node) bool {
	uc.AssertParam("root", root != nil, errors.New("root must not be nil"))

	for _, child := range root.Children {
		ok := simplify_ast(child)
		if ok {
			return true
		}
	}

	idx := -1

	for i := 0; i < len(root.Children)-1; i++ {
		first := root.Children[i]
		second := root.Children[i+1]

		if first.Kind == TextNode && second.Kind == TextNode {
			idx = i
			break
		}
	}

	if idx == -1 {
		return false
	}

	root.Children[idx+1].Data = root.Children[idx].Data + root.Children[idx+1].Data

	root.Children = slices.Delete(root.Children, idx, idx+1)

	return true
}

// PrintAST is a debug function to print the AST.
//
// Parameters:
//   - node: The root node of the tree.
//
// Returns:
//   - string: The string representation of the AST.
//
// Assertions:
//   - Printing the tree should never fail.
func PrintAST(node *Node) string {
	if node == nil {
		return ""
	}

	str, err := fstr.PrintTree(node)
	uc.AssertErr(err, "tree.PrintTree(%s)", node.String())

	return str
}
