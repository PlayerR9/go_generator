package pkg

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"text/template"

	uc "github.com/PlayerR9/MyGoLib/Units/common"
)

// InitLogger initializes the logger with the given prefix.
//
// Parameters:
//   - prefix: The prefix to use for the logger.
//
// Returns:
//   - *log.Logger: The initialized logger. Never nil.
//
// If the prefix is empty, it defaults to "go_generator".
func InitLogger(prefix string) *log.Logger {
	if prefix == "" {
		prefix = "go_generator"
	}

	var builder strings.Builder

	builder.WriteRune('[')
	builder.WriteString(prefix)
	builder.WriteString("]: ")

	logger_prefix := builder.String()

	logger := log.New(os.Stdout, logger_prefix, log.Lshortfile)
	return logger
}

// align_generics is a helper function that aligns the generics in the given StructFieldsVal and GenericsSignVal.
//
// Parameters:
//   - fv: The StructFieldsVal.
//   - gv: The GenericsSignVal.
//
// Returns:
//   - error: An error if the alignment fails (i.e., either the StructFieldsVal or the GenericsSignVal is nil when the other is not).
func align_generics(fv *StructFieldsVal, tv *TypeListVal, gv *GenericsSignVal) error {
	if gv == nil {
		if fv != nil && tv != nil {
			return uc.NewErrInvalidUsage(
				errors.New("not specified the *StructFieldsVal and *TypeListVal but specified the *GenericsSignVal"),
				"Make sure to call go_generator.SetStructFieldsFlag() and go_generator.SetTypeListFlag() as well",
			)
		}
	} else {
		if fv == nil && tv == nil {
			return uc.NewErrInvalidUsage(
				errors.New("not specified the *StructFieldsVal and *TypeListVal but specified the *GenericsSignVal"),
				"Make sure to call go_generator.SetStructFieldsFlag() and go_generator.SetTypeListFlag() as well",
			)
		}
	}

	var all_generics []rune

	if fv != nil {
		for generic_id := range fv.generics {
			pos, ok := slices.BinarySearch(all_generics, generic_id)
			if ok {
				continue
			}

			all_generics = slices.Insert(all_generics, pos, generic_id)
		}
	}

	if tv != nil {
		for generic_id := range tv.generics {
			pos, ok := slices.BinarySearch(all_generics, generic_id)
			if ok {
				continue
			}

			all_generics = slices.Insert(all_generics, pos, generic_id)
		}
	}

	for _, generic_id := range all_generics {
		pos, ok := slices.BinarySearch(gv.letters, generic_id)
		if ok {
			continue
		}

		gv.letters = slices.Insert(gv.letters, pos, generic_id)
		gv.types = slices.Insert(gv.types, pos, "any")
	}

	return nil
}

// ParseFlags parses the command line flags.
//
// Returns:
//   - error: An error if any.
func ParseFlags() error {
	flag.Parse()

	err := align_generics(StructFieldsFlag, TypeListFlag, GenericsSigFlag)
	if err != nil {
		return err
	}

	return nil
}

// Generater is the interface that all generators must implement.
type Generater interface {
	// SetPackageName sets the package name for the generated code.
	//
	// Parameters:
	//   - pkg_name: The package name to use for the generated code.
	//
	// Returns:
	//   - Generater: The same instance of the Generater. Never nil and of the same type as the caller.
	SetPackageName(pkg_name string) Generater
}

// Generate generates code using the given generator and writes it to the given destination file.
//
// WARNING:
//   - Remember to call this function iff the function go_generator.SetOutputFlag() was called
//     and only after the function flag.Parse() was called.
//   - output_loc is the result of the FixOutputLoc() function.
//
// Parameters:
//   - output_loc: The location of the output file.
//   - data: The data to use for the generated code.
//   - t: The template to use for the generated code.
//   - doFunc: Functions to perform on the data before generating the code.
//
// Returns:
//   - error: An error if occurred.
//
// Errors:
//   - *common.ErrInvalidParameter: If any of the parameters is nil or if the actual_loc is an empty string when the
//     IsOutputLocRequired flag was set and the output location was not defined.
//   - error: Any other type of error that may have occurred.
func Generate[T Generater](output_loc string, data T, t *template.Template, doFunc ...func(*T) error) error {
	if t == nil {
		return uc.NewErrNilParameter("t")
	}

	pkg_name, err := FixImportDir(output_loc)
	if err != nil {
		return fmt.Errorf("failed to fix import path: %w", err)
	}

	tmp := data.SetPackageName(pkg_name)
	if tmp == nil {
		return uc.NewErrNilParameter("data")
	}

	data, ok := tmp.(T)
	if !ok {
		return uc.NewErrInvalidParameter("data", uc.NewErrUnexpectedType("data", tmp))
	}

	for _, f := range doFunc {
		if f == nil {
			continue
		}

		err := f(&data)
		if err != nil {
			return err
		}
	}

	var buff bytes.Buffer

	err = t.Execute(&buff, data)
	if err != nil {
		return err
	}

	res := buff.Bytes()

	err = os.WriteFile(output_loc, res, 0644)
	if err != nil {
		return err
	}

	return nil
}
