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

var FREEMEMLOC = 16

// Build the SymbolTable object with known knowns
func generateSymbolTable() map[string]string {
	// Some symbols we already know e.g. @KBD, @SCREEN
	var symbolTable = map[string]string{
		"KBD":    "24576",
		"SCREEN": "16384",
	}

	// Store R1-R15 in symbol table as addresses 1-15
	for i := 0; i < 16; i++ {
		symbolTable[fmt.Sprintf("R%d", i)] = fmt.Sprintf("%d", i)
	}

	return symbolTable
}

// Read a line and determine if it is Symbol, storing and removing if it is
func buildSymbolTable(symbolTable *map[string]string, line string, linenum int) error {

	// Find labels e.g (LABEL) signified by parentheses
	// Store in table as line num of next instruction
	if len(line) > 0 {
		if line[0] == '(' && line[len(line)-1] == ')' {
			label := line[1 : len(line)-1]
			(*symbolTable)[label] = fmt.Sprintf("%d", linenum+1)
			log.Printf("Storing new label %v with line %v", label, linenum)
		}

		// Find Variables e.g. @VAR
		// We define these as @ proceeded by a string value
		// We auto generate memory location (e.g. next after R15) and store in symbol table
		if line[0] == '@' {
			token := line[1:]
			_, err := strconv.Atoi(token)
			// If it errs we probably found a string
			if err != nil {
				// Only store if doesn't exist ()
				if _, ok := (*symbolTable)[token]; !ok {
					(*symbolTable)[token] = fmt.Sprintf("%d", FREEMEMLOC)
					log.Printf("Storing new variable %v in location %v", token, FREEMEMLOC)
					FREEMEMLOC += 1
				} else {
					log.Printf("Duplicate symbol found %v", token)
				}
			}
		}
	}

	return nil
}

// Take a line and return a version of it without whitepace and comments
func cleanline(line string) (string, error) {

	// Strip trailing comments
	before, _, _ := strings.Cut(line, "//")

	// Trim Trailing whitespace
	stripped := strings.Split(before, " ")

	var tokens []string

	// Omit blank tokens
	for _, t := range stripped {
		if len(t) > 0 {
			tokens = append(tokens, t)
		}
	}
	// Check for empty line
	if len(tokens) == 0 {
		return "", errors.New("empty line")
	}

	token := tokens[0]
	// // Check for label lines and remove (their location has already been stored in symbol table)
	if token[0] == '(' && token[len(token)-1] == ')' {
		return token, errors.New("label line")
	}

	return token, nil
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

	// At this point all labels, comments etc. should have been removed (by Cleanlines)
	// We are only dealing with A or C instruction

	if lineAsm[0] == '(' {
		*line = ""
	} else if lineAsm[0] == '@' {
		/*
		 * A Instructions
		 */

		token := lineAsm[1:]
		symbol := (*symbols)[token]
		if symbol != "" {
			// Found symbol, use lookup instead
			token = symbol
		} // Else treating as number

		// See if we have a number
		num, err := strconv.Atoi(token)
		if err != nil {
			// We must have a variable, so look it up
			addr := (*symbols)[token]
			addri, err := strconv.Atoi(addr)
			num = addri
			if err != nil {
				log.Fatalf("Tried to lookup symbol %v, Failed. %v", lineAsm, err)
			}
		}
		*line = fmt.Sprintf("%016b", num) // Convert to base 2, pad with zeroes
	} else {
		/*
		 * C Instructions
		 * dest = comp ; jump
		 */

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

	// Scan all lines, clean, classify and add to symbol table
	// Store cleaned lines in var lines (only instructions)
	var lines []string
	linenum := 1
	for scanner.Scan() {
		line := scanner.Text()

		// Preprocess
		cleaned, err1 := cleanline(line)
		err2 := buildSymbolTable(&symbolTable, cleaned, linenum)
		if err1 != nil || err2 != nil {
			log.Printf("Error in preprocessing: %v %v", err1, err2)
		} else {
			// Increment linenum if A or C instruction
			linenum += 1
			lines = append(lines, cleaned)
		}
	}

	// Translate each line
	var output []string
	for _, line := range lines {
		translate(&line, &symbolTable) // Translate in place
		output = append(output, line)
	}

	// Open output file for writing
	filenameo := "output.hack"
	ofile, err := os.Create(filenameo)
	check(err)
	defer ofile.Close()

	// Write each line token as a line in the output file
	w := bufio.NewWriter(ofile)
	var newline string
	for linenum, t := range output {
		// Omit newline if last line of file or if empty line
		if (linenum != len(output)-1) && len(t) > 0 {
			newline = "\n"
		} else {
			newline = ""
		}
		line := fmt.Sprintf("%v%v", t, newline)
		_, err = w.WriteString(line)
		check(err)
	}
	log.Println("Output to", filenameo)
	w.Flush()
}
