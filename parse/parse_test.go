package parse_test

import (
	"testing"

	"github.com/mmcloughlin/ssarules/ast"
	"github.com/mmcloughlin/ssarules/parse"
)

func TestAdhoc(t *testing.T) {
	src := `
(EqPtr  x x) => (ConstBool [true])
(Slicemask (Const32 [0]))          => (Const32 [0])
(AddPtr <t> x (Const32 [c])) => (OffPtr <t> x [int64(c)])
(EqPtr  (Addr {x} _) (Addr {y} _)) => (ConstBool [x == y])
(EqPtr  (Const(32|64) [c]) (Const(32|64) [d])) => (ConstBool [c == d])
(Not (Less(64|32|16|8) x y)) => (Leq(64|32|16|8) y x)
`

	f, err := parse.String(src)
	if err != nil {
		t.Fatal(err)
	}

	// Debug.
	ast.Print(f)
}
