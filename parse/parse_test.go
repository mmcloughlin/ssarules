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
(Neg32F (Const32F [c])) && c != 0 => (Const32F [-c])
(Trunc64to8  (And64 (Const64 [y]) x)) && y&0xFF == 0xFF => (Trunc64to8 x)
(ZeroExt8to64  (Trunc64to8  x:(Rsh64Ux64 _ (Const64 [s])))) && s >= 56 => x
`

	f, err := parse.String(src)
	if err != nil {
		t.Fatal(err)
	}

	// Debug.
	ast.Print(f)
}
