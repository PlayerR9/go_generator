package parsing

import (
	"testing"
)

func TestLexer(t *testing.T) {
	tokens, err := Lex("{{ .A }} {{ .B }} my_test")
	if err != nil {
		t.Errorf("expected no error, got %s", err.Error())
	}

	if len(tokens) != 10 {
		t.Errorf("expected 4 tokens, got %d", len(tokens))
	}

	if tokens[0].Type != TkOpCurly {
		t.Errorf("expected open curly, got %s", tokens[0].Type)
	}

	if tokens[1].Type != TkDot {
		t.Errorf("expected dot, got %s", tokens[1].Type)
	}

	if tokens[2].Type != TkVariableName {
		t.Errorf("expected variable name, got %s", tokens[2].Type)
	}

	if tokens[3].Type != TkClCurly {
		t.Errorf("expected close curly, got %s", tokens[3].Type)
	}

	if tokens[4].Type != TkOpCurly {
		t.Errorf("expected open curly, got %s", tokens[4].Type)
	}

	if tokens[5].Type != TkDot {
		t.Errorf("expected dot, got %s", tokens[5].Type)
	}

	if tokens[6].Type != TkVariableName {
		t.Errorf("expected variable name, got %s", tokens[6].Type)
	}

	if tokens[7].Type != TkClCurly {
		t.Errorf("expected close curly, got %s", tokens[7].Type)
	}

	if tokens[8].Type != TkText {
		t.Errorf("expected text, got %s", tokens[8].Type)
	}

	if tokens[8].Data != "my_test" {
		t.Errorf("expected \"my_test\", got %s", tokens[8].Data)
	}

	if tokens[9].Type != TkEOF {
		t.Errorf("expected end of file, got %s", tokens[9].Type)
	}
}
