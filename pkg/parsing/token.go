package parsing

type TokenType int

const (
	// Lexer tokens

	// TkEOF is the end of file token.
	TkEOF TokenType = iota

	// TkDot is the dot token.
	TkDot

	// TkNewline is the newline token.
	TkNewline

	// TkText is the text token.
	TkText

	// TkOpCurly is the open curly token.
	TkOpCurly

	// TkClCurly is the close curly token.
	TkClCurly

	// TkVariableName is the variable name token.
	TkVariableName

	// TkWs is the whitespace token.
	TkWs

	// Parsing tokens

	// TkSource is the source token.
	TkSource

	// TkSource1 is the source1 token.
	TkSource1

	// TkElem is the element token.
	TkElem

	// TkVariable is the variable token.
	TkVariable

	// TkSws is the skippable whitespace token.
	TkSws
)

// IsAcceptSymbol implements the parsing.TokenTyper interface.
func (t TokenType) IsAcceptSymbol() bool {
	return t == TkEOF
}

// String implements the parsing.TokenTyper interface.
func (t TokenType) String() string {
	return [...]string{
		"End of File",
		"dot",
		"newline",
		"text",
		"open curly",
		"close curly",
		"variable name",
		"whitespace",

		"Source",
		"Source (I)",
		"Element",
		"Variable",
		"Skippable whitespace",
	}[t]
}
