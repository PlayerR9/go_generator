package parsing

import (
	uc "github.com/PlayerR9/MyGoLib/Units/common"
	uttr "github.com/PlayerR9/go_generator/util/tree"
)

func PrintTokenTree[T uc.Enumer](root *Token[T]) string {
	if root == nil {
		return ""
	}

	str, err := uttr.PrintTree(root)
	uc.AssertErr(err, "tree.PrintTree(%s)", root.GoString())

	return str

}
