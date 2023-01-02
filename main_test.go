package main

import (
	"fmt"
	"testing"
)

/* TODO

- Test line counting (it's zero indexed)
- Test storing labels
-

*/

func TestCInstructions(t *testing.T) {
	// Setup
	var tests = map[string]string{
		"MD=A-1;JGE": "1110110010011011",
		"A-1":        "1110110010000000",
	}

	symbols := generateSymbolTable()

	for instruction, want := range tests {
		// Test
		line := NewLine(instruction)
		line.lineNum = 1
		// Only one line so do simultaneous first and second pass
		updateSymbolTable(&symbols, line)
		line.Translate(&symbols)

		// Assert
		if want != line.translated {
			t.Fatalf(`Expected Translate("%v") = %q, got %q`, instruction, want, line.translated)
		}
	}
}

func TestAInstruction(t *testing.T) {
	// Setup
	instruction := "@4"
	want := "0000000000000100"

	symbols := generateSymbolTable()

	// Test
	line := NewLine(instruction)
	line.lineNum = 1
	// Only one line so do simultaneous first and second pass
	updateSymbolTable(&symbols, line)
	line.Translate(&symbols)

	// Assert
	if want != line.translated {
		t.Fatalf(`Expected Translate("%v") = %q, got %q`, instruction, want, line.translated)
	}
}

func TestVariableInstruction(t *testing.T) {
	// Setup
	var tests = map[string]string{
		"@LABEL": "0000010000000001",
		// "@LABEL":  "0000000000000010", // Using a custom variable
	}

	symbols := generateSymbolTable()

	for instruction, want := range tests {
		// Test
		line := NewLine(instruction)
		line.lineNum = 1024
		// Only one line so do simultaneous first and second pass
		updateSymbolTable(&symbols, line)
		line.Translate(&symbols)

		// Assert
		if want != line.translated {
			t.Fatalf(`Expected Translate("%v") = %q, got %q`, instruction, want, line.translated)
		}
	}
}

func TestBuiltinSymbolsInstruction(t *testing.T) {
	// Setup
	// Instruction -> expected binary
	var tests = map[string]string{
		"@R1":     "0000000000000001",
		"@R15":    "0000000000001111",
		"@KBD":    "0110000000000000", // Using built-in variables
		"@SCREEN": "0100000000000000",
	}

	// Build symbols with our custom label
	symbols := generateSymbolTable()
	// buildSymbolTable(&symbols, "(LABEL)", 1)

	for instruction, want := range tests {
		// Test
		line := NewLine(instruction)
		line.lineNum = 4
		// Only one line so do simultaneous first and second pass
		updateSymbolTable(&symbols, line)
		line.Translate(&symbols)

		// Assert
		if want != line.translated {
			t.Fatalf(`Expected Translate("%v") = %q, got %q`, instruction, want, line.translated)
		}
	}
}

// Test that the input lines gets classified and processes as expected
func TestNewLine(t *testing.T) {

	// Setup
	// rawline: [token, instructionType]
	var tests = map[string][]string{
		"  MD=A-1 // Testing":   {"MD=A-1", "C"},
		"  MD = A-1 // Testing": {"MD=A-1", "C"},
		"  (LABEL) ":            {"LABEL", "L"},
		"  @1 ":                 {"1", "A"},
		" ":                     {"", ""},
		"// comment":            {"", ""},
	}

	for raw, want := range tests {
		// Test
		line := NewLine(raw)
		log := fmt.Sprintf(`Tested %q Got token:%q type:%q`, raw, line.token, line.instructionType)

		// Assert
		if want[0] != line.token || want[1] != line.instructionType {
			t.Fatalf("%v, Wanted token:%q type:%q", log, want[0], want[1])
		} else {
			fmt.Println(log)
		}
	}
}

func TestEmptyLine(t *testing.T) {
	// Setup
	instruction := " // Emptyline"
	want := ""
	symbols := generateSymbolTable()

	// Test
	line := NewLine(instruction)
	line.lineNum = 1024
	updateSymbolTable(&symbols, line)

	// Assert
	if want != line.stripped {
		t.Fatalf(`Cleaned %q expected:%q got %q`, instruction, want, line.stripped)
	}
}
