package main

import (
	"testing"
)

func TestCInstructions(t *testing.T) {
	// Setup
	var tests = map[string]string{
		"MD=A-1;JGE": "1110110010011011",
		"A-1":        "1110110010000000",
	}

	symbols := generateSymbolTable()
	buildSymbolTable(&symbols, "", 1)

	for instruction, want := range tests {
		// Test
		ref := instruction
		translate(&ref, &symbols)

		// Assert
		if want != ref {
			t.Fatalf(`Expected Translate("%v") = %q, got %q`, instruction, want, ref)
		}
	}
}

func TestAInstruction(t *testing.T) {
	// Setup
	instruction := "@4"
	ref := instruction
	want := "0000000000000100"

	symbols := generateSymbolTable()
	buildSymbolTable(&symbols, "", 1)

	// Test
	translate(&ref, &symbols)

	// Assert
	if want != ref {
		t.Fatalf(`Expected Translate("%v") = %q, got %q`, instruction, want, ref)
	}
}

func TestVariableInstruction(t *testing.T) {
	// Setup
	var tests = map[string]string{
		"@LABEL": "0000010000000001",
	}

	symbols := generateSymbolTable()
	buildSymbolTable(&symbols, "(LABEL)", 1024)

	for instruction, want := range tests {
		// Test
		ref := instruction
		translate(&ref, &symbols)

		// Assert
		if want != ref {
			t.Fatalf(`Expected Translate("%v") = %q, got %q`, instruction, want, ref)
		}
	}
}

func TestSymbolicInstruction(t *testing.T) {
	// Setup
	// Instruction -> expected binary
	var tests = map[string]string{
		"@R1":     "0000000000000001",
		"@KBD":    "0110000000000000", // Using built-in variables
		"@SCREEN": "0100000000000000",
		"@LABEL":  "0000000000000010", // Using a custom variable
		// "(LABEL)": "",	// Should produce error (not an instruction)
	}

	// Build symbols with our custom label
	symbols := generateSymbolTable()
	buildSymbolTable(&symbols, "(LABEL)", 1)

	for instruction, want := range tests {
		// Test
		ref := instruction
		translate(&ref, &symbols)

		// Assert
		if want != ref {
			t.Fatalf(`Expected Translate("%v") = %q, got %q`, instruction, want, ref)
		}
	}
}

func TestCleanline(t *testing.T) {
	// Setup
	// instruction := "  MD=A-1 // Testing"
	// want := "MD=A-1"

	var tests = map[string]string{
		"  MD=A-1 // Testing": "MD=A-1",
		// "  (LABEL) ":          "(LABEL)",
		"  @1 ": "@1",
	}

	for instruction, want := range tests {
		// Test
		result, err := cleanline(instruction)

		// Assert
		if want != result || err != nil {
			t.Fatalf(`Expected Cleanline("%v") = %q, got %q, %q`, instruction, want, result, err)
		}
	}
}

func TestEmptyLine(t *testing.T) {
	// Setup
	instruction := " // Emptyline"
	want := ""

	// Test
	result, err := cleanline(instruction)

	// Assert
	if want != result && err.Error() == "empty line" {
		t.Fatalf(`Expected Cleanline("%v") = %q, got %q, %q`, instruction, want, result, err)
	}
}
