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
