package parsing

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	uc "github.com/PlayerR9/MyGoLib/Units/common"
	utpx "github.com/PlayerR9/go_generator/util/parsing"
)

type Lexer struct {
	chars  []rune
	at     int
	tokens []*utpx.Token[TokenType]
}

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

func (l *Lexer) next() (rune, bool) {
	if l.at >= len(l.chars) {
		return utf8.RuneError, false
	}

	first := l.chars[l.at]
	l.at++

	return first, true
}

func (l *Lexer) peek() (rune, bool) {
	if l.at >= len(l.chars) {
		return utf8.RuneError, false
	}

	return l.chars[l.at], true
}

// word = "A".."Z" { "a".."z" } .
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

// variable_name = word { word } .
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

// text = %c { %c } .
func (l *Lexer) lex_text() bool {
	var builder strings.Builder

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
	case ' ', '\t':
		// ws = " " | "\t" . -> skip
		// do nothing
	case '\n':
		// newline = "\n" .
		tk = utpx.NewToken(TkNewline, "\n", nil)
	case '\r':
		// newline = "\r" "\n" .
		next, ok := l.next()
		if !ok {
			return fmt.Errorf("expected '\\n' after '\\r', got nothing instead")
		} else if next != '\n' {
			return fmt.Errorf("expected '\\n' after '\\r', got '\\%c' instead", next)
		}

		tk = utpx.NewToken(TkNewline, "\n", nil)
	case '{':
		// op_curly = "{{" .
		next, ok := l.next()
		if !ok {
			tk = utpx.NewToken(TkText, "{", nil)
		} else if next == '{' {
			tk = utpx.NewToken(TkOpCurly, "{{", nil)
		} else {
			ok := l.lex_text()
			if !ok {
				return fmt.Errorf("unexpected character '%c'", next)
			}
		}
	case '}':
		// cl_curly = "}}" .
		next, ok := l.next()
		if !ok {
			tk = utpx.NewToken(TkText, "}", nil)
		} else if next == '}' {
			tk = utpx.NewToken(TkClCurly, "}}", nil)
		} else {
			ok := l.lex_text()
			if !ok {
				return fmt.Errorf("unexpected character '%c'", next)
			}
		}
	default:
		ok := l.lex_variable_name()
		if !ok {
			ok = l.lex_text()
			if !ok {
				return fmt.Errorf("unexpected character '%c'", curr)
			}
		}
	}

	l.next() // consume

	if tk != nil {
		l.tokens = append(l.tokens, tk)
	}

	return nil
}

func (l *Lexer) is_done() bool {
	return l.at >= len(l.chars)
}

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
