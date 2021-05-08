package parse_test

import (
	"testing"

	"github.com/mmcloughlin/ssarules/ast"
	"github.com/mmcloughlin/ssarules/parse"
)

func TestAdhoc(t *testing.T) {
	src := "(EqPtr  x x) => (ConstBool [true])\n"
	// src := "(Slicemask (Const32 [0]))          => (Const32 [0])\n"
	// src := "(Add x y) => (Add y x)\n"
	f, err := parse.String(src)
	if err != nil {
		t.Fatal(err)
	}

	// Debug.
	ast.Print(f)
}
