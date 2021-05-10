package pass

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mmcloughlin/ssarules/ast"
	"github.com/mmcloughlin/ssarules/parse"
)

func TestExpandAlternates(t *testing.T) {
	cases := []struct {
		Name   string
		Input  []string
		Expect []string
	}{
		{
			Name:   "no_alt",
			Input:  []string{`(Add64 x y) => (Add64 y x)`},
			Expect: []string{`(Add64 x y) => (Add64 y x)`},
		},
		{
			Name: "alt2",
			Input: []string{
				`(Mul(32|64)F x (Const(32|64)F [1])) => x`,
			},
			Expect: []string{
				`(Mul32F x (Const32F [1])) => x`,
				`(Mul64F x (Const64F [1])) => x`,
			},
		},
		{
			Name: "alt4",
			Input: []string{
				`(Not (Less(64|32|16|8) x y)) => (Leq(64|32|16|8) y x)`,
			},
			Expect: []string{
				`(Not (Less64 x y)) => (Leq64 y x)`,
				`(Not (Less32 x y)) => (Leq32 y x)`,
				`(Not (Less16 x y)) => (Leq16 y x)`,
				`(Not (Less8  x y)) => (Leq8  y x)`,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := ParseLines(t, c.Input)
			if err := ExpandAlternates(got); err != nil {
				t.Fatal(err)
			}

			expect := ParseLines(t, c.Expect)

			if diff := cmp.Diff(expect, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func ParseLines(t *testing.T, lines []string) *ast.File {
	t.Helper()
	src := strings.Join(lines, "\n") + "\n"
	f, err := parse.String(src)
	if err != nil {
		t.Fatal(err)
	}
	return f
}
