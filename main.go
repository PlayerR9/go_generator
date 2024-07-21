package main

import (
	"fmt"

	prx "github.com/PlayerR9/go_generator/parsing"
	utpx "github.com/PlayerR9/go_generator/util/parsing"
)

func main() {
	tokens, err := prx.Lex("{{ .A }} {{ .B }} my_type")
	if err != nil {
		fmt.Println(err)
		return
	}

	root, err := prx.Parse(tokens)
	if err != nil {
		fmt.Println(err)
		return
	}

	str := utpx.PrintTokenTree(root)

	fmt.Println(str)
}
