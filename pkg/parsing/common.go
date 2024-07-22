package parsing

import (
	"log"
	"os"

	uc "github.com/PlayerR9/MyGoLib/Units/common"
	utpx "github.com/PlayerR9/go_generator/util/parsing"
)

var (
	// Logger is the logger. Never nil.
	Logger *log.Logger

	// DebugMode is the debug mode. Default is false.
	DebugMode bool
)

func init() {
	Logger = log.New(os.Stdout, "[parsing]: ", log.Lshortfile)

	DebugMode = false
}

// Source = Elem { Elem } EOF .
// Elem = Variable | text | Sws .
// Variable = op_curly [ Sws ] dot variable_name [ Sws ] cl_curly .
// Sws = ws { ws } .

const (
	Grammar string = `
Source = Source1 EOF .
Source1 = Elem .
Source1 = Elem Source1 .
Elem = Variable .
Elem = text .
Elem = Sws .
Variable = op_curly dot variable_name cl_curly .
Variable = op_curly Sws dot variable_name cl_curly .
Variable = op_curly dot variable_name Sws cl_curly .
Variable = op_curly Sws dot variable_name Sws cl_curly .
Sws = ws .
Sws = ws Sws .
`
)

var (
	// SttFunc is the string to type function. Never nil.
	SttFunc utpx.StringToTypeFunc[TokenType]

	// DecisionTable is the decision table. Never nil.
	DecisionTable *utpx.DecisionTable[TokenType]
)

func init() {
	SttFunc = func(field string) (TokenType, bool) {
		switch field {
		case "EOF":
			return TkEOF, true
		case "dot":
			return TkDot, true
		case "newline":
			return TkNewline, true
		case "text":
			return TkText, true
		case "op_curly":
			return TkOpCurly, true
		case "cl_curly":
			return TkClCurly, true
		case "variable_name":
			return TkVariableName, true
		case "ws":
			return TkWs, true
		case "Source":
			return TkSource, true
		case "Source1":
			return TkSource1, true
		case "Elem":
			return TkElem, true
		case "Variable":
			return TkVariable, true
		case "Sws":
			return TkSws, true
		default:
			return 0, false
		}
	}

	dt, err := utpx.NewDecisionTable(Grammar, SttFunc)
	uc.AssertErr(err, "parsing.NewDecisionTable(Grammar, SttFunc)")

	DecisionTable = dt
}
