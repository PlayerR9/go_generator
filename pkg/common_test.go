package go_generator

import (
	"testing"
)

func TestIsValidName(t *testing.T) {
	err := IsValidName("tn", []string{"child"}, NotExported)
	if err != nil {
		t.Errorf("IsValidName failed: %s", err.Error())
	}
}

func TestFixImportDir(t *testing.T) {
	fixed, err := FixImportDir("stack.go")
	if err != nil {
		t.Errorf("FixImportDir failed: %s", err.Error())
	}

	if fixed != "stack.go" {
		t.Errorf("FixImportDir failed: expected %s, got %s", "stack.go", fixed)
	}
}
