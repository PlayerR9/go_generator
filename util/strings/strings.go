package strings

import "strings"

// ConnectWords connects words in a string.
//
// Parameters:
//   - str: The string.
//
// Returns:
//   - string: The string with words connected by underscores.
//
// Examples:
//   - "Hello World" -> "Hello_World"
func ConnectWords(str string) string {
	str = strings.ReplaceAll(str, " ", "_")
	str = strings.ReplaceAll(str, "\t", "_")

	return str
}
