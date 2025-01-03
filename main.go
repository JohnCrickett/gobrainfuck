package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

const defaultCellCount uint = 30000

func main() {
	numCells := flag.Uint("cells", defaultCellCount, "Number of cells to use")
	flag.Parse()
	sourceFiles := flag.Args()

	if len(sourceFiles) > 1 {
		fmt.Fprint(os.Stderr, "Only one source file can be specified\n")
		os.Exit(1)
	}

	if len(sourceFiles) == 0 {
		repl(*numCells)
	} else {
		source, err := os.ReadFile(sourceFiles[0])
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		bytecode, err := compile(string(source))
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		} else {
			execute(bytecode, *numCells, os.Stdin, os.Stdout)
		}
	}
}

func repl(numCells uint) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("BF> ")
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		if line == "exit" || line == "quit" {
			os.Exit(0)
		}
		bc, err := compile(line)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			continue
		}
		execute(bc, numCells, os.Stdin, os.Stdout)

	}
}

type Instruction struct {
	op      uint8
	operand uint16
}

const (
	OpIncrementDp = iota
	OpDecrementDp
	OpIncrement
	OpDecrement
	OpWrite
	OpRead
	OpJumpForward
	OpJumpBackward
)

func compile(source string) ([]Instruction, error) {
	bytecode := make([]Instruction, 0)
	var jmpPositions = make([]uint16, 0)
	var pos uint16
	programSize := len(source)

	for i := 0; i < programSize; i++ {
		cmd := source[i]
		if cmd == '[' {
			bytecode = append(bytecode, Instruction{OpJumpForward, 0})
			jmpPositions = append(jmpPositions, pos)
		} else if cmd == ']' {
			if len(jmpPositions) == 0 {
				return nil, fmt.Errorf("Unmatched jump back at position %d", i)
			}
			jmpTarget := jmpPositions[len(jmpPositions)-1]
			jmpPositions = jmpPositions[:len(jmpPositions)-1]
			bytecode = append(bytecode, Instruction{OpJumpBackward, jmpTarget})
			bytecode[jmpTarget].operand = pos
		} else {
			cmdPos := i
			for i < programSize && cmd == source[i] {
				i++
			}
			count := uint16(i - cmdPos)
			i-- // don't consume mismatching instruction
			switch cmd {
			case '>':
				bytecode = append(bytecode, Instruction{OpIncrementDp, count})
			case '<':
				bytecode = append(bytecode, Instruction{OpDecrementDp, count})
			case '+':
				bytecode = append(bytecode, Instruction{OpIncrement, count})
			case '-':
				bytecode = append(bytecode, Instruction{OpDecrement, count})
			case '.':
				bytecode = append(bytecode, Instruction{OpWrite, count})
			case ',':
				bytecode = append(bytecode, Instruction{OpRead, count})
			default:
				pos-- // don't increment (negate) position for non-instructions
			}
		}
		pos++ // increment position
	}

	if len(jmpPositions) > 0 {
		return nil, fmt.Errorf("Unmatched jump forward at position %d\n", jmpPositions[0])
	}

	return bytecode, nil
}

func execute(bytecode []Instruction, numCells uint, in io.Reader, out io.Writer) {
	cells := make([]uint8, numCells)
	var dp uint
	reader := bufio.NewReader(in)

	for pc := 0; pc < len(bytecode); pc++ {
		switch bytecode[pc].op {
		case OpIncrementDp:
			dp += uint(bytecode[pc].operand)
			if dp >= numCells {
				panic("Access violation, dp out of bounds")
			}
		case OpDecrementDp:
			dp -= uint(bytecode[pc].operand)
			if dp < 0 {
				panic("Access violation, dp out of bounds")
			}
		case OpIncrement:
			cells[dp] += uint8(bytecode[pc].operand)
		case OpDecrement:
			cells[dp] -= uint8(bytecode[pc].operand)
		case OpWrite:
			for r := 0; r < int(bytecode[pc].operand); r++ {
				fmt.Fprintf(out, "%c", cells[dp])
			}
		case OpRead:
			for r := 0; r < int(bytecode[pc].operand); r++ {
				v, _ := reader.ReadByte()
				cells[dp] = v
			}
		case OpJumpForward:
			if cells[dp] == 0 {
				pc = int(bytecode[pc].operand)
			}
		case OpJumpBackward:
			if cells[dp] > 0 {
				pc = int(bytecode[pc].operand)
			}
		default:
			panic(fmt.Sprintf("Unrecognized op %d", bytecode[pc].op))
		}
	}
}
