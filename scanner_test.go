package main

import (
	"slices"
	"testing"
)

type FindReturnsTest struct {
	name string
	in   string
	out  []int
}

var findReturnsTests = []FindReturnsTest{
	// --- Aucun vrai return ---
	{
		name: "empty input",
		in:   "",
		out:  nil,
	},
	{
		name: "inline comment",
		in:   "// return <h1>nope</h1>",
		out:  nil,
	},
	{
		name: "block comment",
		in:   "/* return <h1>nope</h1> */",
		out:  nil,
	},
	{
		name: "double quoted string",
		in:   `s := "return"`,
		out:  nil,
	},
	{
		name: "escaped quote in string",
		in:   `s := "hello \" return world"`,
		out:  nil,
	},
	{
		name: "raw string",
		in:   "s := `return\nreturn`",
		out:  nil,
	},
	{
		name: "identifier prefix",
		in:   "returnValue := 1",
		out:  nil,
	},
	{
		name: "identifier suffix",
		in:   "my_return := 1",
		out:  nil,
	},
	{
		name: "identifier middle",
		in:   "no_return_here := 1",
		out:  nil,
	},

	// --- Vrais returns avec leurs positions (offsets en octets) ---
	{
		name: "at start of string",
		in:   "return",
		out:  []int{0},
	},
	{
		name: "after space and before space",
		in:   "func f() { return 1 }",
		//     012345678901
		out: []int{11},
	},
	{
		name: "after tab and before newline",
		in:   "{\n\treturn\n}",
		//     0 1 2345678
		out: []int{3},
	},
	{
		name: "after semicolon and before paren",
		in:   "x := 1;return(x)",
		//     01234567
		out: []int{7},
	},
	{
		name: "afterinline comment on next line",
		in:   "// fake return\nreturn 1",
		//     012345678901234 5
		out: []int{15},
	},
	{
		name: "after block comment",
		in:   "/* fake return */ return 1",
		//     0123456789012345678
		out: []int{18},
	},
	{
		name: "after string containing return",
		in:   `s := "return"; return s`,
		//     0123456789012345
		out: []int{15},
	},
	{
		name: "after raw string containing return",
		in:   "s := `return`; return s",
		//     0123456789012345
		out: []int{15},
	},
	{
		name: "after rune literal with quote",
		in:   `r := '"'; return r`,
		//     01234567890
		out: []int{10},
	},
	{
		name: "after escaped backslash at end of string",
		in:   `s := "path\\"; return s`,
		//     0123456789012345
		out: []int{15},
	},
	{
		name: "multiple returns",
		in:   "if ok { return 1 }\nreturn 2",
		//     012345678          19
		out: []int{8, 19},
	},
}

func collectAllReturns(in string) []int {
	var positions []int
	offset := 0
	for offset < len(in) {
		pos, found := findReturns(in[offset:])
		if !found {
			break
		}
		positions = append(positions, offset+pos-6)
		offset += pos
	}
	return positions
}

func TestFindReturns(t *testing.T) {
	for _, tt := range findReturnsTests {
		t.Run(tt.name, func(t *testing.T) {
			result := collectAllReturns(tt.in)
			if !slices.Equal(result, tt.out) {
				t.Errorf("failed %s:\ngot:  %v\nwant: %v", tt.name, result, tt.out)
			}
		})
	}
}
