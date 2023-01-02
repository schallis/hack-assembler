# Hack Assembler (from Nand2Tetris Part I Unit 6)

This is my implementation of an Assembler for the Hack computer. This Assembler is written in Go and consistutes my first real Go program.

## Usage

	> go build main.go
	> ./main input.asm
	(outputs to output.hack)
	

## General Strategy:

We use a symbol table to store symbols like Labels e.g.`(LOOP)`, and Variables, both built-in e.g. `KBD` and `SCREEN` as well as user defined e.g. `@i`. All of these jump to specific location (line number or memory location). Note that these variables can be used before their definition (forward references) so we must support this by perfoming two passes on the code, one two build the symbol table, and a second to do the actual translation.

Variables find and reference an unallocated memory location. Labels reference a particular line number and act as a `GOTO`.

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