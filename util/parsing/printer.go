package parsing

import (
	fstr "github.com/PlayerR9/MyGoLib/Formatting/Strings"
	uc "github.com/PlayerR9/lib_units/common"
)

func PrintTokenTree[T TokenTyper](root *Token[T]) string {
	if root == nil {
		return ""
	}

	str, err := fstr.PrintTree(root)
	uc.AssertErr(err, "tree.PrintTree(%s)", root.GoString())

	return str

}
