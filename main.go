package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
)

const defaultCellCount uint = 30000
const Profile = false

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
			//dumpSrc(bytecode)
			execute(bytecode, *numCells, os.Stdin, os.Stdout)
		}
	}
}

func repl(numCells uint) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("CCBF> ")
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
	OpZero
	OpLoopMoveDpLeft
	OpLoopMoveDpRight
	OpLoopMoveValLeft
	OpLoopMoveValRight
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
			loopStart := jmpPositions[len(jmpPositions)-1]
			jmpPositions = jmpPositions[:len(jmpPositions)-1]

			// detect if a loop that can be optimised
			loopSize := pos - loopStart

			if loopSize == 2 && bytecode[pos-1].op == OpDecrement { // [-]
				bytecode = bytecode[:len(bytecode)-2]
				pos -= 2
				bytecode = append(bytecode, Instruction{OpZero, 0})
			} else if loopSize == 2 && bytecode[pos-1].op == OpIncrementDp {
				repeats := bytecode[pos-1].operand
				bytecode = bytecode[:len(bytecode)-2]
				pos -= 2
				bytecode = append(bytecode, Instruction{OpLoopMoveDpRight, repeats})
			} else if loopSize == 2 && bytecode[pos-1].op == OpIncrementDp {
				repeats := bytecode[pos-1].operand
				bytecode = bytecode[:len(bytecode)-2]
				pos -= 2
				bytecode = append(bytecode, Instruction{OpLoopMoveDpLeft, repeats})
			} else if loopSize == 5 &&
				bytecode[loopStart+1].op == OpDecrement && bytecode[loopStart+1].operand == 1 &&
				bytecode[loopStart+3].op == OpIncrement && bytecode[loopStart+3].operand == 1 &&
				bytecode[loopStart+2].op == OpDecrementDp && bytecode[loopStart+4].op == OpIncrementDp &&
				bytecode[loopStart+2].operand == bytecode[loopStart+4].operand {
				// handle [-{1}<{n}+{1}>{n}]
				n := bytecode[loopStart+2].operand
				bytecode = bytecode[:len(bytecode)-5]
				pos -= 5
				bytecode = append(bytecode, Instruction{OpLoopMoveValLeft, n})
			} else if loopSize == 5 &&
				bytecode[loopStart+1].op == OpDecrement && bytecode[loopStart+1].operand == 1 &&
				bytecode[loopStart+3].op == OpIncrement && bytecode[loopStart+3].operand == 1 &&
				bytecode[loopStart+2].op == OpIncrementDp && bytecode[loopStart+4].op == OpDecrementDp &&
				bytecode[loopStart+2].operand == bytecode[loopStart+4].operand {
				// handle [-{1}>{n}+{1}<{n}]
				n := bytecode[loopStart+2].operand
				bytecode = bytecode[:len(bytecode)-5]
				pos -= 5
				bytecode = append(bytecode, Instruction{OpLoopMoveValRight, n})
			} else {
				bytecode = append(bytecode, Instruction{OpJumpBackward, loopStart})
				bytecode[loopStart].operand = pos
			}
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

	loops := make(map[string]int)

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
			start := pc
			if cells[dp] == 0 {
				pc = int(bytecode[pc].operand)
				if Profile {
					loop := loopSource(bytecode, uint16(start), uint16(pc))
					loops[loop] += 1
				}
			}
		case OpJumpBackward:
			if cells[dp] > 0 {
				pc = int(bytecode[pc].operand)
			}
		case OpZero:
			cells[dp] = 0
		case OpLoopMoveDpRight:
			for cells[dp] != 0 {
				dp += uint(bytecode[pc].operand)
			}
		case OpLoopMoveDpLeft:
			for cells[dp] != 0 {
				dp -= uint(bytecode[pc].operand)
			}
		case OpLoopMoveValLeft:
			if cells[dp] != 0 {
				cells[dp-uint(bytecode[pc].operand)] += cells[dp]
				cells[dp] = 0
			}
		case OpLoopMoveValRight:
			if cells[dp] != 0 {
				cells[dp+uint(bytecode[pc].operand)] += cells[dp]
				cells[dp] = 0
			}
		default:
			panic(fmt.Sprintf("Unrecognized op %d", bytecode[pc].op))
		}
	}

	if Profile {
		keys := make([]string, 0, len(loops))
		for key := range loops {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool { return loops[keys[i]] > loops[keys[j]] })

		for i, key := range keys {
			fmt.Printf("%s, %d\n", key, loops[key])
			if i > 10 {
				return
			}
		}
	}

}

func loopSource(bytecode []Instruction, start uint16, end uint16) string {
	ops := make(map[uint8]rune)
	ops[OpIncrementDp] = '>'
	ops[OpDecrementDp] = '<'
	ops[OpIncrement] = '+'
	ops[OpDecrement] = '-'
	ops[OpWrite] = '.'
	ops[OpRead] = ','
	ops[OpJumpForward] = '['
	ops[OpJumpBackward] = ']'

	var loop string

	for i := start; i <= end; i++ {
		if bytecode[i].op == OpJumpForward || bytecode[i].op == OpJumpBackward {
			loop += fmt.Sprintf("%c", ops[bytecode[i].op])
		} else {
			loop += fmt.Sprintf("%c{%d}", ops[bytecode[i].op], bytecode[i].operand)
		}
	}
	return loop
}
