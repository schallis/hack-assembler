package main

import (
	"fmt"
	"strconv"
	"testing"
)

/* TODO

- Test line counting (it's zero indexed)

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

func TestVariables(t *testing.T) {
	// Setup
	var tests = map[string]string{
		// labels should be assigned free spaces incrementally
		"@variable1":   "0000000000010000",
		"@NOTVARIABLE": "shouldnottranslate", // Labels are uppercase
		"@variable2":   "0000000000010001",
	}

	symbols := generateSymbolTable()

	// Test
	for instruction, want := range tests {
		line := NewLine(instruction)
		line.lineNum = 1025
		err := updateSymbolTable(&symbols, line)
		if err == nil {
			line.Translate(&symbols)
			// Assert Valid
			if want != line.translated {
				t.Fatalf(`Expected Translate("%v") = %q, got %q`, instruction, want, line.translated)
			}
		} else {
			// Assert Invalid
			if err.Error() != "invalid" {
				t.Fatalf("Invalid Variable not throwing error")
			}
		}

	}
}

func TestLabel(t *testing.T) {
	// Setup
	var tests = map[string][]string{
		// label ref, label def, translation, linenum
		"@LABEL": {"(LABEL)", "0000010000000000", "1024"},
	}

	symbols := generateSymbolTable()

	// Test
	for instruction, want := range tests {
		// Mock the definition of a label
		preline := NewLine(want[0])
		lineAsInt, _ := strconv.Atoi(want[2])
		preline.lineNum = lineAsInt
		updateSymbolTable(&symbols, preline)

		// Now test the next line which references it
		line := NewLine(instruction)
		line.lineNum = 1025
		updateSymbolTable(&symbols, line)
		line.Translate(&symbols)

		// Assert
		if want[1] != line.translated {
			t.Fatalf(`Expected Translate("%v") = %q, got %q`, instruction, want[1], line.translated)
		}
	}
}

func TestAInstructionsInstruction(t *testing.T) {
	// Setup
	// Instruction -> expected binary
	var tests = map[string]string{
		"@R1":     "0000000000000001",
		"@R15":    "0000000000001111",
		"@KBD":    "0110000000000000", // Using built-in variables
		"@SCREEN": "0100000000000000",
		"@4":      "0000000000000100", // Test raw line numbers
		"@16":     "0000000000010000",
		"@54":     "0000000000110110",
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
