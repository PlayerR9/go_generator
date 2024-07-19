package go_generator

import (
	"errors"
	"fmt"
	"go/build"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	uc "github.com/PlayerR9/MyGoLib/Units/common"
)

var (
	// GoReservedKeywords is a list of Go reserved keywords.
	GoReservedKeywords []string
)

func init() {
	GoReservedKeywords = []string{
		"break", "case", "chan", "const", "continue", "default", "defer", "else",
		"fallthrough", "for", "func", "go", "goto", "if", "import", "interface",
		"map", "package", "range", "return", "select", "struct", "switch", "type",
		"var",
	}
}

// IsGenericsID checks if the input string is a valid single upper case letter and returns it as a rune.
//
// Parameters:
//   - id: The id to check.
//
// Returns:
//   - rune: The valid single upper case letter.
//   - error: An error of type *ErrInvalidID if the input string is not a valid identifier.
func IsGenericsID(id string) (rune, error) {
	if id == "" {
		return '\000', NewErrInvalidID(id, uc.NewErrEmpty(id))
	}

	size := utf8.RuneCountInString(id)
	if size > 1 {
		return '\000', NewErrInvalidID(id, errors.New("value must be a single character"))
	}

	letter := rune(id[0])

	ok := unicode.IsUpper(letter)
	if !ok {
		return '\000', NewErrInvalidID(id, errors.New("value must be an upper case letter"))
	}

	return letter, nil
}

// ParseGenerics parses a string representing a list of generic types enclosed in square brackets.
//
// Parameters:
//   - str: The string to parse.
//
// Returns:
//   - []rune: An array of runes representing the parsed generic types.
//   - error: An error if the parsing fails.
//
// Errors:
//   - *ErrNotGeneric: The string is not a valid list of generic types.
//   - error: An error if the string is a possibly valid list of generic types but fails to parse.
func ParseGenerics(str string) ([]rune, error) {
	if str == "" {
		return nil, NewErrNotGeneric(uc.NewErrEmpty(str))
	}

	var letters []rune

	ok := strings.HasSuffix(str, "]")
	if ok {
		idx := strings.Index(str, "[")
		if idx == -1 {
			err := errors.New("missing opening square bracket")
			return nil, err
		}

		generic := str[idx+1 : len(str)-1]
		if generic == "" {
			err := errors.New("empty generic type")
			return nil, err
		}

		fields := strings.Split(generic, ",")

		for i, field := range fields {
			letter, err := IsGenericsID(field)
			if err != nil {
				err := uc.NewErrAt(i+1, "field", err)
				return nil, err
			}

			letters = append(letters, letter)
		}
	} else {
		letter, err := IsGenericsID(str)
		if err != nil {
			err := NewErrNotGeneric(err)
			return nil, err
		}

		letters = append(letters, letter)
	}

	return letters, nil
}

// FixImportDir takes a destination string and manipulates it to get the correct import path.
//
// Parameters:
//   - dest: The destination path.
//
// Returns:
//   - string: The correct import path.
//   - error: An error if there is any.
func FixImportDir(dest string) (string, error) {
	if dest == "" {
		dest = "."
	}

	dir := filepath.Dir(dest)
	if dir == "." {
		uc, err := build.ImportDir(".", 0)
		if err != nil {
			return "", err
		}

		return uc.Name, nil
	}

	_, right := filepath.Split(dir)
	return right, nil
}

// MakeTypeSig creates a type signature from a type name and a suffix.
//
// It also adds the generic signature if it exists.
//
// Parameters:
//   - type_name: The name of the type.
//   - suffix: The suffix of the type.
//
// Returns:
//   - string: The type signature.
//   - error: An error if the type signature cannot be created. (i.e., the type name is empty)
func MakeTypeSig(type_name string, suffix string) (string, error) {
	if type_name == "" {
		return "", uc.NewErrInvalidParameter("type_name", uc.NewErrEmpty(type_name))
	}

	var builder strings.Builder

	builder.WriteString(type_name)
	builder.WriteString(suffix)

	if GenericsSigFlag == nil {
		return builder.String(), nil
	}

	if len(GenericsSigFlag.letters) > 0 {
		str := GenericsSigFlag.GetSignature()
		builder.WriteString(str)
	}

	return builder.String(), nil
}

// FixOutputLoc fixes the output location.
//
// Parameters:
//   - type_name: The name of the type.
//   - suffix: The suffix of the type.
//
// Returns:
//   - string: The output location.
//   - error: An error if any.
//
// Errors:
//   - *common.ErrInvalidParameter: If the type name is empty.
//   - *common.ErrInvalidUsage: If the OutputLoc flag was not set.
//   - error: Any other error that may have occurred.
func FixOutputLoc(type_name, suffix string) (string, error) {
	output_loc, err := GetOutputLoc()
	if err != nil {
		return "", err
	}

	if type_name == "" {
		return "", uc.NewErrInvalidParameter("type_name", uc.NewErrEmpty(type_name))
	}

	var filename string

	if output_loc == "" {
		var builder strings.Builder

		str := strings.ToLower(type_name)
		builder.WriteString(str)
		builder.WriteString(suffix)

		filename = builder.String()
	} else {
		filename = output_loc
	}

	if output_loc == "" {
		if IsOutputLocRequiredFlag {
			return "", errors.New("flag must be set")
		}

		output_loc = filename
	}

	ext := filepath.Ext(output_loc)
	if ext == "" {
		return "", errors.New("location cannot be a directory")
	} else if ext != ".go" {
		return "", errors.New("location must be a .go file")
	}

	return output_loc, nil
}

// GoExport is an enum that represents whether a variable is exported or not.
type GoExport int

const (
	// NotExported represents a variable that is not exported.
	NotExported GoExport = iota

	// Exported represents a variable that is exported.
	Exported

	// Either represents a variable that is either exported or not exported.
	Either
)

// IsValidName checks if the given variable name is valid.
//
// This function checks if the variable name is not empty and if it is not a
// Go reserved keyword. It also checks if the variable name is not in the list
// of keywords.
//
// Parameters:
//   - variable_name: The variable name to check.
//   - keywords: The list of keywords to check against.
//   - exported: Whether the variable is exported or not.
//
// Returns:
//   - error: An error if the variable name is invalid.
//
// If the variable is exported, the function checks if the variable name starts
// with an uppercase letter. If the variable is not exported, the function checks
// if the variable name starts with a lowercase letter. Any other case, the
// function does not perform any checks.
func IsValidName(variable_name string, keywords []string, exported GoExport) error {
	if variable_name == "" {
		err := uc.NewErrEmpty(variable_name)
		return err
	}

	switch exported {
	case NotExported:
		r, _ := utf8.DecodeRuneInString(variable_name)
		if r == utf8.RuneError {
			return errors.New("invalid UTF-8 encoding")
		}

		ok := unicode.IsLower(r)
		if !ok {
			return errors.New("identifier must start with a lowercase letter")
		}

		ok = slices.Contains(GoReservedKeywords, variable_name)
		if ok {
			err := errors.New("name is a reserved keyword")
			return err
		}
	case Exported:
		r, _ := utf8.DecodeRuneInString(variable_name)
		if r == utf8.RuneError {
			return errors.New("invalid UTF-8 encoding")
		}

		ok := unicode.IsUpper(r)
		if !ok {
			return errors.New("identifier must start with an uppercase letter")
		}
	}

	ok := slices.Contains(keywords, variable_name)
	if ok {
		err := errors.New("name is not allowed")
		return err
	}

	return nil
}

// FixVariableName acts in the same way as IsValidName but fixes the variable name if it is invalid.
//
// Parameters:
//   - variable_name: The variable name to check.
//   - keywords: The list of keywords to check against.
//   - exported: Whether the variable is exported or not.
//
// Returns:
//   - string: The fixed variable name.
//   - error: An error if the variable name is invalid.
func FixVariableName(variable_name string, keywords []string, exported GoExport) (string, error) {
	if variable_name == "" {
		err := uc.NewErrEmpty(variable_name)
		return "", err
	}

	switch exported {
	case NotExported:
		r, size := utf8.DecodeRuneInString(variable_name)
		if r == utf8.RuneError {
			return "", errors.New("invalid UTF-8 encoding")
		}

		if !unicode.IsLetter(r) {
			return "", errors.New("identifier must start with a letter")
		}

		ok := unicode.IsLower(r)
		if !ok {
			r = unicode.ToLower(r)
			variable_name = variable_name[size:]

			var builder strings.Builder

			builder.WriteRune(r)
			builder.WriteString(variable_name)

			variable_name = builder.String()
		}

		ok = slices.Contains(GoReservedKeywords, variable_name)
		if ok {
			return "", fmt.Errorf("variable (%q) is a reserved keyword", variable_name)
		}

		return variable_name, nil
	case Exported:
		r, size := utf8.DecodeRuneInString(variable_name)
		if r == utf8.RuneError {
			return "", errors.New("invalid UTF-8 encoding")
		}

		if !unicode.IsLetter(r) {
			return "", errors.New("identifier must start with a letter")
		}

		ok := unicode.IsUpper(r)
		if !ok {
			r = unicode.ToUpper(r)
			variable_name = variable_name[size:]

			var builder strings.Builder

			builder.WriteRune(r)
			builder.WriteString(variable_name)

			variable_name = builder.String()
		}

		return variable_name, nil
	}

	ok := slices.Contains(keywords, variable_name)
	if ok {
		return "", fmt.Errorf("variable (%q) is already used", variable_name)
	}

	return variable_name, nil
}

// MakeParameterList makes a string representing a list of parameters.
//
// WARNING: Call this function only if StructFieldsFlag is set.
//
// Parameters:
//   - fields: A map of field names and their types.
//
// Returns:
//   - string: A string representing the parameters.
//   - error: An error if any.
func MakeParameterList() (string, error) {
	if StructFieldsFlag == nil {
		return "", uc.NewErrInvalidUsage(
			errors.New("cannot make parameter list without StructFieldsFlag"),
			"Make sure to set StructFieldsFlag before calling this function",
		)
	}

	var field_list []string
	var type_list []string

	for k, v := range StructFieldsFlag.fields {
		if k == "" {
			err := errors.New("found type name with empty name")
			return "", err
		}

		first_letter := rune(k[0])

		ok := unicode.IsLetter(first_letter)
		if !ok {
			err := fmt.Errorf("type name %q must start with a letter", k)
			return "", err
		}

		ok = unicode.IsUpper(first_letter)
		if !ok {
			continue
		}

		pos, ok := slices.BinarySearch(field_list, k)
		uc.AssertF(!ok, "%q must be unique", k)

		field_list = slices.Insert(field_list, pos, k)
		type_list = slices.Insert(type_list, pos, v)
	}

	var values []string
	var builder strings.Builder

	for i := 0; i < len(field_list); i++ {
		param := strings.ToLower(field_list[i])

		builder.WriteString(param)
		builder.WriteRune(' ')
		builder.WriteString(type_list[i])

		str := builder.String()
		values = append(values, str)

		builder.Reset()
	}

	joined_str := strings.Join(values, ", ")

	return joined_str, nil
}

// MakeAssignmentList makes a string representing a list of assignments.
//
// WARNING: Call this function only if StructFieldsFlag is set.
//
// Parameters:
//   - fields: A map of field names and their types.
//
// Returns:
//   - string: A string representing the assignments.
//   - error: An error if any.
func MakeAssignmentList() (map[string]string, error) {
	if StructFieldsFlag == nil {
		return nil, uc.NewErrInvalidUsage(
			errors.New("cannot make assignment list without StructFieldsFlag"),
			"Make sure to set the StructFieldsFlag before calling this function",
		)
	}

	var field_list []string
	var type_list []string

	for k, v := range StructFieldsFlag.fields {
		if k == "" {
			err := errors.New("found type name with empty name")
			return nil, err
		}

		first_letter := rune(k[0])

		ok := unicode.IsLetter(first_letter)
		if !ok {
			err := fmt.Errorf("type name %q must start with a letter", k)
			return nil, err
		}

		ok = unicode.IsUpper(first_letter)
		if !ok {
			continue
		}

		pos, ok := slices.BinarySearch(field_list, k)
		uc.AssertF(!ok, "%q must be unique", k)

		field_list = slices.Insert(field_list, pos, k)
		type_list = slices.Insert(type_list, pos, v)
	}

	assignment_map := make(map[string]string)

	for i := 0; i < len(field_list); i++ {
		param := strings.ToLower(field_list[i])

		assignment_map[field_list[i]] = param
	}

	return assignment_map, nil
}

var (
	// ZeroValueTypes is a list of types that have a default value of zero.
	ZeroValueTypes []string

	// NillablePrefix is a list of prefixes that indicate a type is nillable.
	NillablePrefix []string
)

func init() {
	ZeroValueTypes = []string{
		"byte",
		"complex64",
		"complex128",
		"uint",
		"uint8",
		"uint16",
		"uint32",
		"uint64",
		"uintptr",
		"int",
		"int8",
		"int16",
		"int32",
		"int64",
	}

	NillablePrefix = []string{
		"[]",
		"map",
		"*",
		"chan",
		"func",
		"interface",
		"<-",
	}
}

// ZeroValueOf returns the zero value of a type.
//
// Parameters:
//   - type_name: The name of the type.
//
// Returns:
//   - string: The zero value of the type.
func ZeroValueOf(type_name string) string {
	if type_name == "" {
		return ""
	}

	for _, prefix := range NillablePrefix {
		if strings.HasPrefix(type_name, prefix) {
			return "nil"
		}
	}

	switch type_name {
	case "bool":
		return "false"
	case "error", "any":
		return "nil"
	case "float32", "float64":
		return "0.0"
	case "rune":
		return "'\\u0000'"
	case "string":
		return "\"\""
	}

	ok := slices.Contains(ZeroValueTypes, type_name)
	if ok {
		return "0"
	}

	return "*new(" + type_name + ")"
}
