package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

/*
	General Strategy:

		* Ignore comments (inline and line, indentation)
		* Read line
		* Break line into parts (ignore whitespace, semicolons etc.)
		* Figure out:
			A instruction or C instruction
		* If C instruction
			Lookup binary code for each token part (dest, comp, jmp)
		* If A instruction
			Convert from decimal to binary
			Concatenate (with any required padding bits)

	Symbols like Labels, Variables, KBD, SCREEN Etc.
	e.g. loop, jump to specific location defined earlier
	Support for variables, defined earlier
	Both of these will require a symbol table mapping symbol to address
	Find unallocated memory location, allocate and store
	NOTE: will have to build symbol table in first pass,
	then use second time around to support forward references
*/

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Read a line and determine if it is Symbol, storing and removing if it is
func buildSymbolTable(line string, linenum int) (map[string]string, error) {
	// Some symbols we already know e.g. @KBD, @SCREEN
	var symbols = map[string]string{
		"KBD":    "24576",
		"SCREEN": "16384",
	}

	// Find labels e.g (LABEL) and add to table
	if len(line) > 0 && line[0] == '(' && line[len(line)-1] == ')' {
		label := line[1 : len(line)-1]
		// Label
		symbols[label] = fmt.Sprintf("%d", linenum) // TODO: Set to line number
	}

	// Find Variables e.g. @VAR

	return symbols, nil
}

// Take a line and return a version of it without whitepace and comments
func cleanline(line string) (string, error) {

	// Strip trailing comments
	before, _, _ := strings.Cut(line, "//")

	// Trim Trailing whitespace
	tokens := strings.Split(before, " ")

	var nonblank []string

	// Omit blank tokens
	for _, t := range tokens {
		if len(t) > 0 {
			nonblank = append(nonblank, t)
		}
	}
	// Check for empty line
	if len(nonblank) == 0 {
		return "", errors.New("empty line")
	}

	return nonblank[0], nil
}

// Take a cleaned line and translate it into binary
// e.g. 1110110010011011 -> MD=A-1;JGE
func translate(line *string, symbols *map[string]string) {

	var dmap = map[string]string{
		"null": "000",
		"M":    "001",
		"D":    "010",
		"MD":   "011",
		"A":    "100",
		"AM":   "101",
		"AD":   "110",
		"AMD":  "111",
	}

	var jmap = map[string]string{
		"null": "000",
		"JGT":  "001",
		"JEQ":  "010",
		"JGE":  "011",
		"JLT":  "100",
		"JNE":  "101",
		"JLE":  "110",
		"JMP":  "111",
	}

	var cmap = map[string]string{
		// A=0
		"0":   "0101010",
		"1":   "0111111",
		"-1":  "0111010",
		"D":   "0001100",
		"A":   "0110000",
		"!D":  "0001101",
		"!A":  "0110001",
		"-D":  "0001111",
		"-A":  "0110011",
		"D+1": "0011111",
		"A+1": "0110111",
		"D-1": "0001110",
		"A-1": "0110010",
		"D+A": "0000010",
		"D-A": "0010011",
		"A-D": "0000111",
		"D&A": "0000000",
		"D|A": "0010101",
		// A = 1
		"M":   "1110000",
		"!M":  "1110001",
		"-M":  "1110011",
		"M+1": "1110111",
		"M-1": "1110010",
		"D+M": "1000010",
		"D-M": "1010011",
		"M-D": "1000111",
		"D&M": "1000000",
		"D|M": "1010101",
	}

	lineAsm := *line

	// A this point all labels, comments etc. should have been removed (by Cleanlines)
	// We are only dealing with A or C instruction
	if lineAsm[0] == '@' {
		// A instruction
		number := lineAsm[1:]
		symbol := (*symbols)[number]
		if symbol != "" {
			// Found symbol, use lookup instead
			number = symbol
		} else {
			log.Printf("Tried to look up symbol %v, didn't find. Treating as number", lineAsm)
		}
		num, err := strconv.Atoi(number)
		if err != nil {
			log.Fatalf("Tried to use symbol %v as number, Failed. %v", lineAsm, err)
		}
		*line = fmt.Sprintf("%016b", num) // Convert to base 2, pad with zeroes
	} else {
		// C instruction
		// dest = comp ; jump

		i := 1
		x := 11
		dest := "000"
		comp := "0000000" // will be prefixed with A during lookup
		jump := "000"

		// Determine Jump
		// Split on `;` producing [dest/comp, jump]
		destcomp := comp
		jumpsplit := strings.Split(lineAsm, ";")
		destcomp = jumpsplit[0]
		if len(jumpsplit) > 1 {
			// We have a jump e.g. 0;JMP
			jump = jmap[jumpsplit[1]]
		}

		// Break down comp side
		// Split on `=` producing [dest, comp] or just [comp]
		compsplit := strings.Split(destcomp, "=")
		if len(compsplit) > 1 {
			// we have a destination e.g. A=D+1
			dest = dmap[compsplit[0]]
			comp = cmap[compsplit[1]]
		} else {
			// Just a comp e.g. D+1
			comp = cmap[compsplit[0]]
		}

		// Use lookup tables to determine a, d, j
		*line = fmt.Sprintf("%v%v%v%v%v", i, x, comp, dest, jump)
	}

	log.Printf("%v	%v", *line, lineAsm)
}

// Read a .asm file specified as the only argument
// Assemble and produce a .hack file in the same folder as run
func main() {
	var err error
	log.SetPrefix("debug: ")
	log.SetFlags(0)

	// Read the args for the filename .asm file
	args := os.Args
	filename := ""
	if len(args) < 2 || args[1] == "" {
		log.Printf("No filename specified as first arg. Defaulting to input.asm")
		filename = "input.asm"
	} else {
		filename = args[1]
	}

	// Open file
	file, err := os.Open(filename)
	check(err)
	defer file.Close()

	// Scan through it line by line
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Process each line each line (clean, translate)
	var tokens []string
	for linenum, line := range lines {
		token, err1 := cleanline(line)
		symbols, err2 := buildSymbolTable(line, linenum)
		if err1 == nil && err2 == nil {
			translate(&token, &symbols) // Translate in place
			tokens = append(tokens, token)
		}
	}

	// Open output file for writing
	ofile, err := os.Create("output.hack")
	check(err)
	defer ofile.Close()

	// Write each line token as a line in the output file
	w := bufio.NewWriter(ofile)
	for _, t := range tokens {
		line := fmt.Sprintf("%v\n", t)
		// log.Println(line)
		_, err = w.WriteString(line)
		check(err)
	}
	w.Flush()
}
