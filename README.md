# Hack Assembler (from Nand2Tetris Part I Unit 6)

This is my implementation of an Assembler for the Hack computer to fulfill the requirements of the [Nand2Tetris](https://www.nand2tetris.org/course) course. This Assembler is written in Go and translates Hack Assembly files into machine code that runs on the computer built during the course.

The specification is described [here](https://www.nand2tetris.org/_files/ugd/44046b_89a8e226476741a3b7c5204575b8a0b2.pdf)

## Usage

	> go build main.go
	> ./main input.asm
	(outputs to output.hack)

## Test Suite

	> go test
	

## General Strategy:

We use a symbol table to store symbols such as Labels (e.g.`(LOOP)`) and Variables (both built-in e.g. `KBD` as well as user defined e.g. `@my-variable`). All of these jump to specific location (line number or memory location). Note that these variables can be used before their definition (i.e. forward references) so we must support this by perfoming two passes on the code, one two build the symbol table, and a second to do the actual translation.

Variables are assigned to unallocated memory locations. Labels reference a particular line number and act as a `GOTO`.

### First pass
- Classify all lines, keeping only A and C instruction for further processing
- Store line numbers
- Store labels in symbol table
	- Store non-numeric A intructions as symbols
**NOTE:** will have to build symbol table in first pass,
then use second time around to support forward references

### Second pass
- Lookup all A instructions to find symbols
- Translate each instruction
	- If C instruction
		- Lookup binary code for each token part (dest, comp, jmp)
	- If A instruction
		- Convert from decimal to binary
		- Concatenate (with any required padding bits)