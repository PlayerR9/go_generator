package parsing

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	utpx "github.com/PlayerR9/go_generator/util/parsing"
	uc "github.com/PlayerR9/lib_units/common"
)

// Lexer is a lexical analyzer for the template.
type Lexer struct {
	// chars is the input stream.
	chars []rune

	// at is the current position in the input stream.
	at int

	// tokens is the list of tokens.
	tokens []*utpx.Token[TokenType]
}

// set_input_stream is a helper function that sets the input stream.
//
// Parameters:
//   - str: The input stream.
//
// Returns:
//   - error: An error if the input stream is invalid.
//
// Errors:
//   - *common.ErrInvalidParameter: If str is empty.
//   - *common.ErrAt: If the utf-8 encoding is invalid at a specific position.
func (l *Lexer) set_input_stream(str string) error {
	if len(str) == 0 {
		return uc.NewErrInvalidParameter("str", uc.NewErrEmpty(str))
	}

	var chars []rune

	for i := 0; len(str) > 0; i++ {
		c, size := utf8.DecodeRuneInString(str)
		if c == utf8.RuneError {
			return uc.NewErrAt(i+1, "char", errors.New("invalid utf-8 encoding"))
		}

		chars = append(chars, c)
		str = str[size:]
	}

	l.chars = chars

	return nil
}

// next is a helper function that returns the next rune in the input stream.
//
// Returns:
//   - rune: The next rune in the input stream.
//   - bool: True if the next rune is valid, false otherwise.
//
// utf8.RuneError is returned whenever the function returns false.
func (l *Lexer) next() (rune, bool) {
	if l.at >= len(l.chars) {
		return utf8.RuneError, false
	}

	first := l.chars[l.at]
	l.at++

	return first, true
}

// peek is a helper function that returns the next rune in the input stream without consuming it.
//
// Returns:
//   - rune: The next rune in the input stream.
//   - bool: True if the next rune is valid, false otherwise.
//
// utf8.RuneError is returned whenever the function returns false.
func (l *Lexer) peek() (rune, bool) {
	if l.at >= len(l.chars) {
		return utf8.RuneError, false
	}

	return l.chars[l.at], true
}

// lex_word is a helper function that lexes a word.
//
// Here's the EBNF rule for a word:
//
//	word = "A".."Z" { "a".."z" } .
//
// Returns:
//   - string: The word.
//   - bool: True if the word is valid, false otherwise.
func (l *Lexer) lex_word() (string, bool) {
	curr, ok := l.peek()
	if !ok || !unicode.IsUpper(curr) {
		return "", false
	}

	l.next() // consume

	var builder strings.Builder

	builder.WriteRune(curr)

	for {
		next, ok := l.peek()
		if !ok {
			break
		} else if !unicode.IsLower(next) {
			break
		}

		builder.WriteRune(next)

		l.next() // consume
	}

	return builder.String(), true
}

// lex_variable_name is a helper function that lexes a variable name.
//
// Here's the EBNF rule for a variable name:
//
//	variable_name = word { word } .
//
// Returns:
//   - bool: True if the variable name is valid, false otherwise.
func (l *Lexer) lex_variable_name() bool {
	var builder strings.Builder

	for {
		word, ok := l.lex_word()
		if !ok {
			break
		}

		builder.WriteString(word)
	}

	if builder.Len() == 0 {
		return false
	}

	tk := utpx.NewToken(TkVariableName, builder.String(), nil)

	l.tokens = append(l.tokens, tk)

	return true
}

// lex_text is a helper function that lexes a text.
//
// Here's the EBNF rule for a text:
//
//	text = %c { %c } .
//
// Parameters:
//   - prev: The previous rune. (if any)
//
// Returns:
//   - bool: True if the text is valid, false otherwise.
func (l *Lexer) lex_text(prev *rune) bool {
	var builder strings.Builder

	if prev != nil {
		builder.WriteRune(*prev)
	}

	for {
		next, ok := l.peek()
		if !ok || next == '{' {
			break
		}

		builder.WriteRune(next)

		l.next() // consume
	}

	if builder.Len() == 0 {
		return false
	}

	tk := utpx.NewToken(TkText, builder.String(), nil)

	l.tokens = append(l.tokens, tk)

	return true
}

// lex_one is a helper function that lexes a single token.
//
// Returns:
//   - error: An error if the token is invalid, nil otherwise.
func (l *Lexer) lex_one() error {
	curr, ok := l.peek()
	if !ok {
		return fmt.Errorf("unexpected end of input")
	}

	var tk *utpx.Token[TokenType]

	switch curr {
	case '.':
		// dot = "." .
		tk = utpx.NewToken(TkDot, ".", nil)

		l.next() // consume
	case ' ', '\t':
		// ws = " " | "\t" .
		tk = utpx.NewToken(TkWs, string(curr), nil)

		l.next() // consume
	case '\n':
		// newline = "\n" .
		tk = utpx.NewToken(TkNewline, "\n", nil)

		l.next() // consume
	case '\r':
		l.next() // consume

		// newline = "\r" "\n" .
		next, ok := l.next()
		if !ok {
			return fmt.Errorf("expected '\\n' after '\\r', got nothing instead")
		} else if next != '\n' {
			return fmt.Errorf("expected '\\n' after '\\r', got '\\%c' instead", next)
		}

		tk = utpx.NewToken(TkNewline, "\n", nil)
	case '{':
		l.next() // consume

		// op_curly = "{{" .
		next, ok := l.peek()
		if !ok {
			tk = utpx.NewToken(TkText, "{", nil)
		} else if next == '{' {
			tk = utpx.NewToken(TkOpCurly, "{{", nil)

			l.next() // consume
		} else {
			ok := l.lex_text(&curr)
			if !ok {
				return fmt.Errorf("unexpected character '%c'", next)
			}
		}
	case '}':
		l.next() // consume

		// cl_curly = "}}" .
		next, ok := l.peek()
		if !ok {
			tk = utpx.NewToken(TkText, "}", nil)
		} else if next == '}' {
			tk = utpx.NewToken(TkClCurly, "}}", nil)

			l.next() // consume
		} else {
			ok := l.lex_text(&curr)
			if !ok {
				return fmt.Errorf("unexpected character '%c'", next)
			}
		}
	default:
		ok := l.lex_variable_name()
		if !ok {
			ok = l.lex_text(nil)
			if !ok {
				return fmt.Errorf("unexpected character '%c'", curr)
			}
		}
	}

	if tk != nil {
		l.tokens = append(l.tokens, tk)
	}

	return nil
}

// is_done is a helper function that checks if the lexer is done.
//
// Returns:
//   - bool: True if the lexer is done, false otherwise.
func (l *Lexer) is_done() bool {
	return l.at >= len(l.chars)
}

// get_tokens is a helper function that returns the tokens.
//
// Returns:
//   - []*utpx.Token[TokenType]: The tokens.
//
// This function adds the EOF token and sets the lookaheads for the tokens.
func (l *Lexer) get_tokens() []*utpx.Token[TokenType] {
	eof := utpx.NewToken(TkEOF, "", nil)

	l.tokens = append(l.tokens, eof)

	for i := 0; i < len(l.tokens)-1; i++ {
		l.tokens[i].Lookahead = l.tokens[i+1]
	}

	return l.tokens
}

func Lex(str string) ([]*utpx.Token[TokenType], error) {
	l := &Lexer{}

	err := l.set_input_stream(str)
	if err != nil {
		return nil, fmt.Errorf("invalid input stream: %s", err.Error())
	}

	for !l.is_done() {
		err := l.lex_one()
		if err != nil {
			tokens := l.get_tokens()
			return tokens, err
		}
	}

	tokens := l.get_tokens()

	return tokens, nil
}
