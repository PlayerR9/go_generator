package pkg

import (
	"errors"
	"flag"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"

	uc "github.com/PlayerR9/MyGoLib/Units/common"
)

var (
	// OutputLocFlag is a flag that specifies the location of the output file.
	OutputLocFlag *string

	// IsOutputLocRequiredFlag is a flag that specifies whether the output location is required or not.
	IsOutputLocRequiredFlag bool

	// StructFieldsFlag is a pointer to the fields_flag flag.
	StructFieldsFlag *StructFieldsVal

	// GenericsSigFlag is a pointer to the generics_flag flag.
	GenericsSigFlag *GenericsSignVal

	// TypeListFlag is a pointer to the type_list_flag flag.
	TypeListFlag *TypeListVal
)

// SetOutputFlag sets the flag that specifies the location of the output file.
//
// Parameters:
//   - def_value: The default value of the output_flag flag.
//   - required: Whether the flag is required or not.
//
// Here are all the possible valid calls to this function:
//
//	SetOutputFlag("", false) <-> SetOutputFlag("[no location]", false)
//	SetOutputFlag("path/to/file.go", false)
//	SetOutputFlag("", true) <-> SetOutputFlag("path/to/file.go", true)
//
// Documentation:
//
// **Flag: Output File**
//
// This optional flag is used to specify the output file. If not specified, the output will be written to
// standard output, that is, the file "<type_name>_treenode.go" in the root of the current directory.
func SetOutputFlag(def_value string, required bool) {
	var usage string

	if required {
		var builder strings.Builder

		builder.WriteString("The location of the output file. ")
		builder.WriteString("It must be set and it must specify a .go file.")

		usage = builder.String()
	} else {
		var def_loc string

		if def_value == "" {
			def_loc = "\"[no location]\""
		} else {
			def_loc = strconv.Quote(def_value)
		}

		var builder strings.Builder

		builder.WriteString("The location of the output file. ")

		builder.WriteString("If set, it must specify a .go file. ")
		builder.WriteString("On the other hand, if not set, the default location of ")
		builder.WriteString(def_loc)
		builder.WriteString(" will be used instead.")

		usage = builder.String()
	}

	OutputLocFlag = flag.String("o", "", usage)
	IsOutputLocRequiredFlag = required
}

// GetOutputLoc gets the location of the output file.
//
// Returns:
//   - string: The location of the output file.
//   - error: An error of type *common.ErrInvalidUsage if the output location was not defined
//     prior to calling this function.
func GetOutputLoc() (string, error) {
	if OutputLocFlag == nil {
		return "", uc.NewErrInvalidUsage(
			errors.New("output location was not defined"),
			"Please call the go_generator.SetOutputFlag() function before calling this function.",
		)
	}

	return *OutputLocFlag, nil
}

// StructFieldsVal is a struct that represents the fields value.
type StructFieldsVal struct {
	// fields is a map of the fields and their types.
	fields map[string]string

	// generics is a map of the generics and their types.
	generics map[rune]string

	// is_required is a flag that specifies whether the fields value is required or not.
	is_required bool

	// count is the number of fields expected. -1 for unlimited number of fields.
	count int
}

// String implements the flag.Value interface.
//
// Format:
//
//	"<value1> <type1>, <value2> <type2>, ..."
func (s *StructFieldsVal) String() string {
	if len(s.fields) == 0 {
		return ""
	}

	var values []string
	var builder strings.Builder

	for name, t := range s.fields {
		builder.WriteString(name)
		builder.WriteRune(' ')
		builder.WriteString(t)

		str := builder.String()
		values = append(values, str)

		builder.Reset()
	}

	joined_str := strings.Join(values, ", ")
	quoted := strconv.Quote(joined_str)

	return quoted
}

// Set implements the flag.Value interface.
func (s *StructFieldsVal) Set(value string) error {
	if value == "" && s.is_required {
		return errors.New("value must be set")
	}

	fields := strings.Split(value, ",")

	parsed := make(map[string]string)

	for i, field := range fields {
		if field == "" {
			continue
		}

		sub_fields := strings.Split(field, "/")

		if len(sub_fields) == 1 {
			reason := errors.New("missing type")
			err := uc.NewErrAt(i+1, "field", reason)
			return err
		} else if len(sub_fields) > 2 {
			reason := errors.New("too many fields")
			err := uc.NewErrAt(i+1, "field", reason)
			return err
		}

		parsed[sub_fields[0]] = sub_fields[1]
	}

	if s.count != -1 && len(parsed) != s.count {
		return fmt.Errorf("wrong number of fields: expected %d, got %d", s.count, len(parsed))
	}

	s.fields = parsed

	// Find generics
	generics := make(map[rune]string)

	for _, field_type := range s.fields {
		chars, err := ParseGenerics(field_type)
		ok := IsErrNotGeneric(err)
		if ok {
			continue
		} else if err != nil {
			err := fmt.Errorf("syntax error for type %q: %w", field_type, err)
			return err
		}

		for _, char := range chars {
			_, ok := generics[char]
			if !ok {
				generics[char] = ""
			}
		}
	}

	return nil
}

// SetStructFieldsFlag sets the flag that specifies the fields of the struct to generate the code for.
//
// Parameters:
//   - flag_name: The name of the flag.
//   - is_required: Whether the flag is required or not.
//   - count: The number of fields expected. -1 for unlimited number of fields.
//   - brief: A brief description of the flag.
//
// Any negative number will be interpreted as unlimited number of fields. Also, the value 0 will not set the flag.
//
// Documentation:
//
// **Flag: Fields**
//
// The "fields" flag is used to specify the fields that the tree node contains. Because it doesn't make
// a lot of sense to have a tree node without fields, this flag must be set.
//
// Its argument is specified as a list of key-value pairs where each pair is separated by a comma (",") and
// a slash ("/") is used to separate the key and the value.
//
// The key indicates the name of the field while the value indicates the type of the field.
//
// For instance, running the following command:
//
//	//go:generate treenode -type="TreeNode" -fields=a/int,b/int,name/string
//
// will generate a tree node with the following fields:
//
//	type TreeNode struct {
//		// Node pointers.
//
//		a int
//		b int
//		name string
//	}
//
// It is important to note that spaces are not allowed.
//
// Also, it is possible to specify generics by following the value with the generics between square brackets;
// like so: "a/MyType[T,C]"
func SetStructFieldsFlag(flag_name string, is_required bool, count int, brief string) {
	if count == 0 {
		return
	}

	if count < 0 {
		count = -1
	}

	StructFieldsFlag = &StructFieldsVal{
		fields:      make(map[string]string),
		generics:    make(map[rune]string),
		is_required: is_required,
		count:       count,
	}

	var usage strings.Builder

	usage.WriteString(brief)

	if is_required {
		if count == -1 {
			usage.WriteString("It must be set with at least one field.")
		} else {
			usage.WriteString(fmt.Sprintf("It must be set with exactly %d fields.", count))
		}
	} else {
		if count == -1 {
			usage.WriteString("It is optional but, if set, it must be set with at least one field.")
		} else {
			usage.WriteString(fmt.Sprintf("It is optional but, if set, it must be set with exactly %d fields.", count))
		}
	}

	usage.WriteString("The syntax of the this flag is described in the documentation.")

	flag.Var(StructFieldsFlag, flag_name, usage.String())
}

// GetFields returns the fields of the struct.
//
// Returns:
//   - map[string]string: A map of field names and their types.
func (s *StructFieldsVal) GetFields() map[string]string {
	return s.fields
}

// GenericsSignVal is a struct that contains the values of the generics.
type GenericsSignVal struct {
	// letters is a slice that contains the letters of the generics.
	letters []rune

	// types is a slice that contains the types of the generics.
	types []string

	// is_required is a flag that specifies whether the generics value is required or not.
	is_required bool

	// count is a flag that specifies the number of generics.
	count int
}

// String implements the flag.Value interface.
//
// Format:
//
//	[letter1 type1, letter2 type2, ...]
func (s *GenericsSignVal) String() string {
	if len(s.letters) == 0 {
		return ""
	}

	var values []string
	var builder strings.Builder

	for i, letter := range s.letters {
		builder.WriteRune(letter)
		builder.WriteRune(' ')
		builder.WriteString(s.types[i])

		str := builder.String()
		values = append(values, str)

		builder.Reset()
	}

	joined_str := strings.Join(values, ", ")

	builder.WriteRune('[')
	builder.WriteString(joined_str)
	builder.WriteRune(']')

	str := builder.String()
	return str
}

// Set implements the flag.Value interface.
func (s *GenericsSignVal) Set(value string) error {
	if value == "" {
		return nil
	}

	fields := strings.Split(value, ",")

	for i, field := range fields {
		if field == "" {
			continue
		}

		letter, g_type, err := parse_generics_value(field)
		if err != nil {
			return uc.NewErrAt(i+1, "field", err)
		}

		err = s.add(letter, g_type)
		if err != nil {
			return uc.NewErrAt(i+1, "field", err)
		}
	}

	if s.count != -1 && len(s.letters) != s.count {
		return fmt.Errorf("invalid number of generics: expected %d, got %d", s.count, len(s.letters))
	}

	return nil
}

// SetGenericsSignFlag sets the flag that specifies the generics to generate the code for.
//
// Parameters:
//   - flag_name: The name of the flag.
//   - is_required: Whether the flag is required or not.
//
// Documentation:
//
// **Flag: Generics**
//
// This optional flag is used to specify the type(s) of the generics. However, this only applies if at least one
// generic type is specified in the fields flag. If none, then this flag is ignored.
//
// As an edge case, if this flag is not specified but the fields flag contains generics, then
// all generics are set to the default value of "any".
//
// As with the fields flag, its argument is specified as a list of key-value pairs where each pair is separated
// by a comma (",") and a slash ("/") is used to separate the key and the value. The key indicates the name of
// the generic and the value indicates the type of the generic.
//
// For instance, running the following command:
//
//	//go:generate treenode -type="TreeNode" -fields=a/MyType[T],b/MyType[C] -g=T/any,C/int
//
// will generate a tree node with the following fields:
//
//	type TreeNode[T any, C int] struct {
//		// Node pointers.
//
//		a T
//		b C
//	}
func SetGenericsSignFlag(flag_name string, is_required bool, count int) {
	if count == 0 {
		return
	}

	if count < 0 {
		count = -1
	}

	GenericsSigFlag = &GenericsSignVal{
		letters:     make([]rune, 0),
		types:       make([]string, 0),
		is_required: is_required,
		count:       count,
	}

	var usage strings.Builder

	usage.WriteString("The signature of generics.")

	if is_required {
		usage.WriteString("It must be set.")
	} else {
		usage.WriteString("It is optional.")
	}

	usage.WriteString("The syntax of the this flag is described in the documentation.")

	flag.Var(GenericsSigFlag, flag_name, usage.String())
}

// parse_generics_value is a helper function that is used to parse the generics
// values.
//
// Parameters:
//   - field: The field to parse.
//
// Returns:
//   - rune: The letter of the generic.
//   - string: The type of the generic.
//   - error: An error if the parsing fails.
//
// Errors:
//   - *ErrInvalidID: If the id is invalid.
//   - error: If the parsing fails.
//
// Assertions:
//   - field != ""
func parse_generics_value(field string) (rune, string, error) {
	uc.Assert(field != "", "field must not be an empty string")

	sub_fields := strings.Split(field, "/")

	if len(sub_fields) == 1 {
		return '\000', "", errors.New("missing type of generic")
	} else if len(sub_fields) > 2 {
		return '\000', "", errors.New("too many fields")
	}

	left := sub_fields[0]

	letter, err := IsGenericsID(left)
	if err != nil {
		return '\000', "", err
	}

	right := sub_fields[1]

	return letter, right, nil
}

// add is a helper function that is used to add a generic to the GenericsValue.
//
// Parameters:
//   - letter: The letter of the generic.
//   - g_type: The type of the generic.
//
// Errors:
//   - error: If the parsing fails.
//
// Assertions:
//   - letter is an upper case letter.
//   - g_type != ""
func (gv *GenericsSignVal) add(letter rune, g_type string) error {
	uc.AssertParam("letter", unicode.IsUpper(letter), errors.New("letter must be an upper case letter"))
	uc.AssertParam("g_type", g_type != "", errors.New("type must be set"))

	pos, ok := slices.BinarySearch(gv.letters, letter)
	if !ok {
		gv.letters = slices.Insert(gv.letters, pos, letter)
		gv.types = slices.Insert(gv.types, pos, g_type)

		return nil
	}

	if gv.types[pos] != g_type {
		err := fmt.Errorf("duplicate definition for generic %q: %s and %s", string(letter), gv.types[pos], g_type)
		return err
	}

	return nil
}

// GetSignature returns the signature of the generics.
//
// Format:
//
//	[T1, T2, T3]
//
// Returns:
//   - string: The list of generics.
func (gv *GenericsSignVal) GetSignature() string {
	if len(gv.letters) == 0 {
		return ""
	}

	values := make([]string, 0, len(gv.letters))

	for _, letter := range gv.letters {
		str := string(letter)
		values = append(values, str)
	}

	joined_str := strings.Join(values, ", ")

	var builder strings.Builder

	builder.WriteRune('[')
	builder.WriteString(joined_str)
	builder.WriteRune(']')

	str := builder.String()

	return str
}

// TypeListVal is a struct that represents a list of types.
type TypeListVal struct {
	// fields is a list of types.
	types []string

	// generics is a map of the generics and their types.
	generics map[rune]string

	// is_required is a flag that specifies whether the fields value is required or not.
	is_required bool

	// count is the number of fields expected.
	count int
}

// String implements the flag.Value interface.
//
// Format:
//
//	"<type1>, <type2>, ..."
func (s *TypeListVal) String() string {
	if len(s.types) == 0 {
		return ""
	}

	joined_str := strings.Join(s.types, ", ")
	quoted := strconv.Quote(joined_str)

	return quoted
}

// Set implements the flag.Value interface.
func (s *TypeListVal) Set(value string) error {
	if value == "" && s.is_required {
		return errors.New("value must be set")
	}

	parsed := strings.Split(value, ",")

	var top int

	for i := 0; i < len(parsed); i++ {
		if parsed[i] != "" {
			parsed[top] = parsed[i]
			top++
		}
	}

	parsed = parsed[:top]

	if s.count != -1 && len(parsed) != s.count {
		return fmt.Errorf("wrong number of types: expected %d, got %d", s.count, len(parsed))
	}

	if s.count != -1 && len(parsed) != s.count {
		return fmt.Errorf("wrong number of fields: expected %d, got %d", s.count, len(parsed))
	}

	s.types = parsed

	// Find generics
	generics := make(map[rune]string)

	for _, field_type := range s.types {
		chars, err := ParseGenerics(field_type)
		ok := IsErrNotGeneric(err)
		if ok {
			continue
		} else if err != nil {
			err := fmt.Errorf("syntax error for type %q: %w", field_type, err)
			return err
		}

		for _, char := range chars {
			_, ok := generics[char]
			if !ok {
				generics[char] = ""
			}
		}
	}

	return nil
}

// SetTypeListFlag sets the flag that specifies the fields of the struct to generate the code for.
//
// Parameters:
//   - flag_name: The name of the flag.
//   - is_required: Whether the flag is required or not.
//   - count: The number of fields expected. -1 for unlimited number of fields.
//   - brief: A brief description of the flag.
//
// Any negative number will be interpreted as unlimited number of fields. Also, the value 0 will not set the flag.
//
// Documentation:
//
// **Flag: Types**
//
// The "types" flag is used to specify a list of types that are accepted by the generator.
//
// Its argument is specidied as a list of Go types separated by commas without spaces.
//
// For instance, running the following command:
//
//	//go:generate table -name=IntTable -type=int -fields=a/int,b/int,name/string
//
// will generate a tree node with the following fields:
//
//	type TreeNode struct {
//		// Node pointers.
//
//		a int
//		b int
//		name string
//	}
//
// It is important to note that spaces are not allowed.
//
// Also, it is possible to specify generics by following the value with the generics between square brackets;
// like so: "a/MyType[T,C]"
func SetTypeListFlag(flag_name string, is_required bool, count int, brief string) {
	if count == 0 {
		return
	}

	if count < 0 {
		count = -1
	}

	TypeListFlag = &TypeListVal{
		types:       make([]string, 0),
		generics:    make(map[rune]string),
		is_required: is_required,
		count:       count,
	}

	var usage strings.Builder

	usage.WriteString(brief)

	if is_required {
		if count == -1 {
			usage.WriteString("It must be set with at least one field.")
		} else {
			usage.WriteString(fmt.Sprintf("It must be set with exactly %d fields.", count))
		}
	} else {
		if count == -1 {
			usage.WriteString("It is optional but, if set, it must be set with at least one field.")
		} else {
			usage.WriteString(fmt.Sprintf("It is optional but, if set, it must be set with exactly %d fields.", count))
		}
	}

	usage.WriteString("The syntax of the this flag is described in the documentation.")

	flag.Var(TypeListFlag, flag_name, usage.String())
}

// GetType returns the type at the given index.
//
// Parameters:
//   - idx: The index of the type to return.
//
// Return:
//   - string: The type at the given index.
//   - error: An error of type *uc.ErrInvalidParameter if the index is out of bounds.
func (s *TypeListVal) GetType(idx int) (string, error) {
	if idx < 0 || idx >= len(s.types) {
		return "", uc.NewErrInvalidParameter("idx", uc.NewErrOutOfBounds(idx, 0, len(s.types)))
	}

	return s.types[idx], nil
}
