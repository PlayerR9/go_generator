package parsing

type TokenType int

const (
	TkEOF TokenType = iota
	TkDot
	TkNewline
	TkText
	TkOpCurly
	TkClCurly
	TkVariableName

	TkSource
	TkSource1
	TkElem
	TkVariable
)

func (t TokenType) String() string {
	return [...]string{
		"End of File",
		"dot",
		"newline",
		"text",
		"open curly",
		"close curly",
		"variable name",

		"Source",
		"Source (I)",
		"Element",
		"Variable",
	}[t]
}
