package pass

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mmcloughlin/ssarules/ast"
	"github.com/mmcloughlin/ssarules/internal/test"
	"github.com/mmcloughlin/ssarules/parse"
)

func TestExpandAlternatesCases(t *testing.T) {
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

func TestExpandAlternatesErrors(t *testing.T) {
	cases := []struct {
		Name           string
		Input          []string
		ErrorSubstring string
	}{
		{
			Name: "mismatch_op_args",
			Input: []string{
				`(Mul(32|64)F x (Const(8|16|32|64)F [1])) => x`,
			},
			ErrorSubstring: "incompatible alternative counts 2 and 4",
		},
		{
			Name: "mismatch_opparts",
			Input: []string{
				`(Mul(8|16|32|64)(L|Q) x) => x`,
			},
			ErrorSubstring: "incompatible alternative counts 4 and 2",
		},
		{
			Name: "mismatch_match_result",
			Input: []string{
				`(Mul(64|32|16|8) ...) => (MUL(Q|L|L) ...)`,
			},
			ErrorSubstring: "incompatible alternative counts 4 and 3",
		},
		{
			Name: "mismatch_value",
			Input: []string{
				`(Opcode (Add(1|2)To(A|B|C) x x) x) => x`,
			},
			ErrorSubstring: "incompatible alternative counts 2 and 3",
		},
		{
			Name: "mismatch_result_only",
			Input: []string{
				`(Opcode x y) => (Mul(8|16|32|64)(L|Q) x y)`,
			},
			ErrorSubstring: "incompatible alternative counts 4 and 2",
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := ParseLines(t, c.Input)
			err := ExpandAlternates(got)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), c.ErrorSubstring) {
				t.Fatalf("expected error to contain %q; got %q", c.ErrorSubstring, err)
			}
		})
	}
}

func TestExpandAlternatesFiles(t *testing.T) {
	test.Glob(t, "../internal/testdata/*.rules", func(t *testing.T, filename string) {
		// Parse file.
		f, err := parse.File(filename)
		if err != nil {
			t.Fatal(err)
		}

		pre := len(f.Rules)

		// Apply ExpandAlternates pass.
		if err := ExpandAlternates(f); err != nil {
			t.Fatal(err)
		}
		post := len(f.Rules)

		t.Logf("expanded %d rules to %d", pre, post)

		if post < pre {
			t.Fatal("fewer rules after expansion")
		}
	})
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
