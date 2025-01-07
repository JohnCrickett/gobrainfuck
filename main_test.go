package main

import (
	"bytes"
	"reflect"
	"testing"
)

func TestCompile(t *testing.T) {
	tests := map[string]struct {
		in       string
		expected []Instruction
	}{
		"Cat": {
			in: ",[.,]",
			expected: []Instruction{
				Instruction{OpRead, 1},
				Instruction{OpJumpForward, 4},
				Instruction{OpWrite, 1},
				Instruction{OpRead, 1},
				Instruction{OpJumpBackward, 1},
			},
		},
		"Optimisation": {
			in: ">>>>>+>+++",
			expected: []Instruction{
				Instruction{OpIncrementDp, 5},
				Instruction{OpIncrement, 1},
				Instruction{OpIncrementDp, 1},
				Instruction{OpIncrement, 3},
			},
		},
		"Optimisation of Loops - Zero Cell": {
			in: "[-]",
			expected: []Instruction{
				Instruction{OpZero, 0},
			},
		},
		"Optimisation of Loops - Zero Cell, Surrounding Code": {
			in: "++[-]++>+<-<",
			expected: []Instruction{
				Instruction{OpIncrement, 2},
				Instruction{OpZero, 0},
				Instruction{OpIncrement, 2},
				Instruction{OpIncrementDp, 1},
				Instruction{OpIncrement, 1},
				Instruction{OpDecrementDp, 1},
				Instruction{OpDecrement, 1},
				Instruction{OpDecrementDp, 1},
			},
		},
		"Optimisation of Loops - Move Dp": {
			in: "++[>>>]--",
			expected: []Instruction{
				Instruction{OpIncrement, 2},
				Instruction{OpLoopMoveDpRight, 3},
				Instruction{OpDecrement, 2},
			},
		},
		"Optimisation of Loops - Move Data Left": {
			in: ">>>>++[-<<<+>>>]",
			expected: []Instruction{
				Instruction{OpIncrementDp, 4},
				Instruction{OpIncrement, 2},
				Instruction{OpLoopMoveValLeft, 3},
			},
		},
		"Optimisation of Loops - Move Data Right": {
			in: ">>>>>++[->>>>+<<<<]--",
			expected: []Instruction{
				Instruction{OpIncrementDp, 5},
				Instruction{OpIncrement, 2},
				Instruction{OpLoopMoveValRight, 4},
				Instruction{OpDecrement, 2},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bc, err := compile(test.in)
			if err != nil {
				t.Fatal(err)
			}
			if len(bc) != len(test.expected) || !reflect.DeepEqual(bc, test.expected) {
				t.Errorf("got %v, want %v", bc, test.expected)
			}
		})
	}
}

func TestCompileMismatchedJumps(t *testing.T) {
	tests := map[string]struct {
		in string
	}{
		"Jump Forward":  {in: "[+>+>[<<]."},
		"Jump Backward": {in: "[+>+>[<<].],+>>]."},
	}
	for name, test := range tests {
		bc, err := compile(test.in)
		if err == nil {
			t.Errorf("Expected error for mismatched jump in test %s", name)
		}
		if bc != nil {
			t.Errorf("Expected nil bytecode when and error occurs in test %s", name)
		}
	}
}

func TestExecute(t *testing.T) {
	tests := map[string]struct {
		in       []Instruction
		expected *bytes.Buffer
	}{
		"Add and print": {
			in: []Instruction{
				Instruction{OpIncrement, 2},
				Instruction{OpWrite, 1},
			},
			expected: bytes.NewBuffer([]byte{2}),
		},
		"Zero": {
			in: []Instruction{
				Instruction{OpIncrement, 2},
				Instruction{OpWrite, 1},
				Instruction{OpZero, 0},
				Instruction{OpWrite, 1},
			},
			expected: bytes.NewBuffer([]byte{2, 0}),
		},
		"Move Data": {
			in: []Instruction{
				Instruction{OpIncrementDp, 2},
				Instruction{OpIncrement, 2},
				Instruction{OpWrite, 1},
				Instruction{OpLoopMoveValLeft, 2},
				Instruction{OpDecrementDp, 2},
				Instruction{OpWrite, 1},
			},
			expected: bytes.NewBuffer([]byte{2, 2}),
		},
		"Move Data Pointer Right": {
			in: []Instruction{
				Instruction{OpIncrementDp, 6},
				Instruction{OpIncrement, 1},
				Instruction{OpWrite, 1},
				Instruction{OpDecrementDp, 6},
				Instruction{OpWrite, 1},
				Instruction{OpLoopMoveDpRight, 2},
				Instruction{OpWrite, 1},
			},
			expected: bytes.NewBuffer([]byte{1, 0, 1}),
		},
		"Move Data Pointer Left": {
			in: []Instruction{
				Instruction{OpIncrement, 1},
				Instruction{OpWrite, 1},
				Instruction{OpIncrementDp, 7},
				Instruction{OpIncrement, 1},
				Instruction{OpWrite, 1},
				Instruction{OpIncrementDp, 1},
				Instruction{OpLoopMoveDpLeft, 2},
				Instruction{OpIncrementDp, 1},
				Instruction{OpWrite, 1},
				Instruction{OpLoopMoveDpRight, 2},
				Instruction{OpWrite, 1},
			},
			expected: bytes.NewBuffer([]byte{1, 1, 0, 1}),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var in bytes.Buffer
			var out bytes.Buffer
			execute(test.in, 300, &in, &out)

			if bytes.Compare(out.Bytes(), test.expected.Bytes()) != 0 {
				t.Errorf("got %v, want %v", out, test.expected)
			}
		})
	}
}
