package main

import (
	"bytes"
	"fmt"

	pkg "github.com/PlayerR9/go_generator/pkg"
)

func main() {
	t, err := pkg.NewTemplate(templ)
	if err != nil {
		fmt.Println(err)
		return
	}

	type MyStruct struct {
		TypeName string
	}

	data := MyStruct{
		TypeName: "Lexer",
	}

	err = t.Apply(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	var buff bytes.Buffer

	err = t.Write(&buff)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(buff.String())
}

const templ string = `{{  .A }} my_type {{ .B }}`
