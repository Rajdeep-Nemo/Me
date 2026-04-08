package evaluator

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Rajdeep-Nemo/sugarglaze/internal/object"
)

// Handles the string formatting, interpolation, and strict type checking.
// It returns the final string to print, or a runtime Error if types mismatch.
func formatOutput(args ...object.Object) (string, *object.Error) {
	if len(args) == 0 {
		return "", nil
	}

	firstArg, isString := args[0].(*object.String)

	// If the first argument isn't a string, or it doesn't contain '{',
	// just squish all arguments together without spaces
	if !isString || !strings.Contains(firstArg.Value, "{") {
		var out strings.Builder
		for _, arg := range args {
			out.WriteString(arg.Inspect())
		}
		return out.String(), nil
	}

	// Parses the format string
	format := firstArg.Value
	var out strings.Builder
	argIndex := 1 // Variables to inject start at args[1]

	i := 0
	for i < len(format) {
		if format[i] == '{' {
			// Find the matching '}'
			closeIdx := strings.IndexByte(format[i:], '}')
			if closeIdx != -1 {
				// Extract the type expected by the placeholder (e.g., "i32" from "{i32}")
				expectedType := format[i+1 : i+closeIdx]

				// Check if enough arguments are provided for the placeholders
				if argIndex >= len(args) {
					return "", &object.Error{Message: "not enough arguments for format string"}
				}

				providedArg := args[argIndex]

				// STRICT TYPE CHECKING: Compare the placeholder to the object's actual type!
				if string(providedArg.Type()) != expectedType {
					return "", &object.Error{
						Message: fmt.Sprintf("type mismatch in print: expected '{%s}', got '%s'", expectedType, providedArg.Type()),
					}
				}

				// If it passes, write the value and move to the next argument
				out.WriteString(providedArg.Inspect())
				argIndex++
				i += closeIdx + 1 // Jump past the '}'
				continue
			}
		}
		// Write normal characters
		out.WriteByte(format[i])
		i++
	}

	// Check if too many arguments are provided for the placeholders
	if argIndex < len(args) {
		return "", &object.Error{Message: "too many arguments for format string"}
	}

	return out.String(), nil
}

// Prints exact string representations
func builtinPrint(args ...object.Object) object.Object {
	output, err := formatOutput(args...)
	if err != nil {
		return err
	}
	fmt.Print(output)
	return &object.Nil{}
}

// Prints with a trailing newline
func builtinPrintln(args ...object.Object) object.Object {
	output, err := formatOutput(args...)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return &object.Nil{}
}

// Trims leading and trailing whitespace from a string
func trim(s string) string {
	return strings.TrimSpace(s)
}

// A shared buffered reader for os.Stdin.
var reader = bufio.NewReader(os.Stdin)

// InputStr reads a line from stdin with an optional prompt.
func InputStr(prompt string) string {
	if prompt != "" {
		fmt.Print(prompt)
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read input.")
		os.Exit(1)
	}
	return strings.TrimRight(line, "\r\n")
}

// InputChar reads a single character from stdin with an optional prompt.
func InputChar(prompt string) byte {
	line := InputStr(prompt)
	if line == "" {
		fmt.Fprintln(os.Stderr, "Empty input.")
		os.Exit(1)
	}
	if len(line) > 1 {
		fmt.Fprintln(os.Stderr, "Multiple characters found.")
		os.Exit(1)
	}
	return line[0]
}

// InputBool reads a boolean value from stdin with an optional prompt (true/false).
func InputBool(prompt string) bool {
	line := trim(InputStr(prompt))
	if line == "" {
		fmt.Fprintln(os.Stderr, "Empty input.")
		os.Exit(1)
	}
	if len(line) > 5 {
		fmt.Fprintf(os.Stderr, "Invalid value: '%s', value accepted: true / false\n", line)
		os.Exit(1)
	}
	if line == "true" {
		return true
	}
	if line == "false" {
		return false
	}
	fmt.Fprintf(os.Stderr, "Invalid value: '%s', value accepted: true / false\n", line)
	os.Exit(1)
	return false
}
