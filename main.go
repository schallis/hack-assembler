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

// The line struct stores information about the lines we are translating
type Line struct {
	raw string

	// computed values (by NewLine constructor)
	stripped        string
	token           string
	empty           bool   // default: false
	instructionType string // `A`, `C` or 'L' for Label

	// computed in first pass
	lineNum int

	// computed in second pass
	translated string
}

// Constructor for the Line type
func NewLine(rawline string) Line {
	line := Line{
		raw: rawline,
	}
	line.clean()
	line.classify()

	return line
}

// Is A instruction?
func (l *Line) isA() bool {
	return l.instructionType == "A"
}

// Is C instruction?
func (l *Line) isC() bool {
	return l.instructionType == "C"
}

// Is Label?
func (l *Line) isL() bool {
	return l.instructionType == "L"
}

func (l *Line) clean() {
	// Strip trailing comments
	before, _, _ := strings.Cut(l.raw, "//")

	// Remove all whitespace
	stripped := strings.Replace(before, " ", "", -1)

	// Check for empty line
	if len(stripped) == 0 {
		l.empty = true
	} else {
		l.stripped = stripped
	}
}

// Test if line is a label and return if it is. Error if not
func (l *Line) getLabel() (string, error) {
	last := len(l.stripped) - 1
	if l.stripped[0] == '(' && l.stripped[last] == ')' {
		return l.stripped[1:last], nil
	}
	return "", errors.New("not a label")
}

// Classify as A, C or L (Label)
// or leave classification nil (e.g. for comments or blank lines)
// Also store raw token e.g "@TOKEN" = "TOKEN" and "(LABEL)" = "LABEL"
func (l *Line) classify() {
	if !l.empty {
		if l.stripped[0] == '@' {
			l.instructionType = "A"
			l.token = l.stripped[1:]
		} else if label, err := l.getLabel(); err == nil {
			l.instructionType = "L"
			l.token = label
		} else {
			l.instructionType = "C"
			l.token = l.stripped
		}
	}
}

// Utility function for error handling
func check(e error) {
	if e != nil {
		panic(e)
	}
}

var FREEMEMLOC = 15

// Build the SymbolTable object with known knowns
func generateSymbolTable() map[string]int {
	// Some symbols we already know e.g. @KBD, @SCREEN
	var symbolTable = map[string]int{
		"KBD":    24576,
		"SCREEN": 16384,
	}

	// Store R1-R15 in symbol table as addresses 1-15
	for i := 0; i < 16; i++ {
		symbolTable[fmt.Sprintf("R%d", i)] = i
	}

	return symbolTable
}

// Read a line and determine if it is Symbol, storing and removing if it is
func updateSymbolTable(symbolTable *map[string]int, line Line) error {

	// Find labels e.g (LABEL) signified by parentheses
	// Store in table as line num of next instruction
	if line.isL() {
		(*symbolTable)[line.token] = line.lineNum
		log.Printf("Storing new label %v with line %v", line.token, line.lineNum)
	}

	// Find Variables e.g. @VAR
	// We define these as @ proceeded by a string value
	// We auto generate memory location (e.g. next after R15) and store in symbol table
	if line.isA() {
		_, err := strconv.Atoi(line.token)
		// If it errs we found a non-numeric string
		if err != nil {
			// Only store if doesn't exist
			if _, ok := (*symbolTable)[line.token]; !ok {
				(*symbolTable)[line.token] = FREEMEMLOC
				log.Printf("Storing new variable %v in location %v", line.token, FREEMEMLOC)
				FREEMEMLOC += 1
			} else {
				log.Printf("Duplicate symbol found %v", line.token)
			}
		}
	}

	return nil
}

// Take a line struct, translate it into binary and store translation
// e.g. MD=A-1;JGE -> 1110110010011011
func (line *Line) Translate(symbols *map[string]int) {

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

	if line.isA() {
		// See if there is a lookup
		number, ok := (*symbols)[line.token]
		if ok {
			// Found symbol so translate
			line.translated = fmt.Sprintf("%016b", number)
		} else {
			// Not found, assume number
			number, err := strconv.Atoi(line.token)
			if err != nil {
				// Not number, must be a missing symbol
				log.Fatalf("Tried to lookup symbol %v, Failed. %v", line.token, err)
			}
			line.translated = fmt.Sprintf("%016b", number)
		}
	} else if line.isC() {
		i := 1
		x := 11
		dest := "000"
		comp := "0000000" // will be prefixed with A during lookup
		jump := "000"

		// Determine Jump
		// Split on `;` producing [dest/comp, jump]
		destcomp := comp
		jumpsplit := strings.Split(line.token, ";")
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
		line.translated = fmt.Sprintf("%v%v%v%v%v", i, x, comp, dest, jump)
	} else {
		// Can only translate A or C instruction
		log.Fatalf("Attempted to translate non-instruction: %q", line.stripped)
	}
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
		// filename = "input.asm"
		filename = "/Users/stevenchallis/Desktop/nand2tetris/projects/06/max/Max.asm"
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

	symbolTable := generateSymbolTable()

	// First Pass
	var processedLines []*Line
	lineNum := 0
	for scanner.Scan() {
		text := scanner.Text()
		inLine := NewLine(text)
		inLine.lineNum = lineNum

		// Store line for second pass with computed line number
		if inLine.isA() || inLine.isC() {
			lineNum += 1
			processedLines = append(processedLines, &inLine)
		}

		// Find any symbols and add them to the table
		err := updateSymbolTable(&symbolTable, inLine)
		if err != nil {
			log.Printf("Error updating symbols: %v", err)
		}
	}

	// Second Pass
	var outLines []*Line
	for _, line := range processedLines {
		line.Translate(&symbolTable)
		outLines = append(outLines, line)
	}

	// Open output file for writing
	filenameo := "output.hack"
	ofile, err := os.Create(filenameo)
	check(err)
	defer ofile.Close()

	// Write each line token as a line in the output file
	w := bufio.NewWriter(ofile)
	var newline string
	for lineNum, t := range outLines {
		// Omit newline if last line of file or if empty line
		if lineNum != len(outLines)-1 {
			newline = "\n"
		} else {
			newline = ""
		}
		DEBUG := false
		var line string
		if DEBUG {
			line = fmt.Sprintf("%-3v %-16v %v%v", t.lineNum, t.stripped, t.translated, newline)
		} else {
			line = fmt.Sprintf("%v%v", t.translated, newline)
		}
		_, err = w.WriteString(line)
		check(err)
	}
	log.Println("Output to", filenameo)
	w.Flush()
}
