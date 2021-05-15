package rules

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/mmcloughlin/ssarules/ast"
	"github.com/mmcloughlin/ssarules/parse"
)

func TestPredicates(t *testing.T) {
	cases := []struct {
		Name      string
		Source    string
		Predicate func(*ast.Rule) bool
		Expect    bool
	}{
		{
			Name:      "binding",
			Source:    `(MOVWUreg x:(MOVBUload _ _)) => (MOVDreg x)`,
			Predicate: HasBinding,
			Expect:    true,
		},
		{
			Name:      "type",
			Source:    `(AddPtr <t> x (Const64 [c])) => (OffPtr <t> [c] x)`,
			Predicate: HasType,
			Expect:    true,
		},
		{
			Name:      "aux",
			Source:    `(EqPtr (Addr {a} _) (Addr {b} _)) => (ConstBool [a == b])`,
			Predicate: HasAux,
			Expect:    true,
		},
		{
			Name:      "ellipsis",
			Source:    `(SignExt16to32 ...) => (MOVHreg ...)`,
			Predicate: HasEllipsis,
			Expect:    true,
		},
		{
			Name:      "trailing",
			Source:    `(SelectN [2] (MakeResult a b c ___)) => c`,
			Predicate: HasTrailing,
			Expect:    true,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			f, err := parse.String(c.Source)
			if err != nil {
				t.Fatal(err)
			}

			if len(f.Rules) != 1 {
				t.Fatalf("expected one rule")
			}
			r := f.Rules[0]

			if got := c.Predicate(r); got != c.Expect {
				t.Logf("source = %s", c.Source)
				t.Logf("predicate = %v", runtime.FuncForPC(reflect.ValueOf(c.Predicate).Pointer()).Name())
				t.Fatalf("got %v; expect %v", got, c.Expect)
			}
		})
	}
}
