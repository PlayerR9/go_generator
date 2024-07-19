package main

import (
	"log"
	"text/template"

	ggen "github.com/PlayerR9/go_generator/pkg"
)

var (
	t      *template.Template
	Logger *log.Logger
)

func init() {
	t = template.Must(template.New("").Parse(templ))
	Logger = ggen.InitLogger("go_generator")
}

var ()

func init() {
	ggen.SetOutputFlag("", false)
}

type GenData struct {
	PackageName string
}

func (g GenData) SetPackageName(pkg_name string) ggen.Generater {
	g.PackageName = pkg_name
	return g
}

func main() {
	err := ggen.ParseFlags()
	if err != nil {
		Logger.Fatalf("Could not parse flags: %s", err.Error())
	}

	output_loc, err := ggen.FixOutputLoc("foo", "foo_suffix")
	if err != nil {
		Logger.Fatalf("Could not fix output location: %s", err.Error())
	}

	err = ggen.Generate(output_loc, GenData{}, t)
	if err != nil {
		Logger.Fatalf("Could not generate code: %s", err.Error())
	}
}

const templ = `
package main

import (
	"log"
	"text/template"

	ggen "github.com/PlayerR9/go_generator/pkg"
)

var (
	t *template.Template
	Logger *log.Logger
)

func init() {
	t = template.Must(template.New("").Parse(templ))
	Logger = ggen.InitLogger("go_generator")
}

var (
	// Put here your flags.
)

func init() {
	ggen.SetOutputFlag("", false)
}

type GenData struct {
	PackageName string
}

func (g GenData) SetPackageName(pkg_name string) ggen.Generater {
	g.PackageName = pkg_name
	return g
}

func main() {
	err := ggen.ParseFlags()
	if err != nil {
		Logger.Fatalf("Could not parse flags: %s", err.Error())
	}

	output_loc, err := ggen.FixOutputLoc("foo", "foo_suffix")
	if err != nil {
		Logger.Fatalf("Could not fix output location: %s", err.Error())
	}

	err = ggen.Generate("foo.go", GenData{}, t)
	if err != nil {
		Logger.Fatalf("Could not generate code: %s", err.Error())
	}
}

const templ = "my template"
`
