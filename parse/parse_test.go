package parse_test

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mmcloughlin/ssarules/ast"
	"github.com/mmcloughlin/ssarules/parse"
)

func TestFiles(t *testing.T) {
	filenames, err := filepath.Glob("testdata/*.rules")
	if err != nil {
		t.Fatal(err)
	}
	for _, filename := range filenames {
		t.Run(filepath.Base(filename), func(t *testing.T) {
			_, err := parse.File(filename)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCases(t *testing.T) {
	cases := []struct {
		Name   string
		Source string
		Expect *ast.Rule
	}{
		{
			// Reference: https://github.com/golang/go/blob/02ce4118219dc51a14680a0c5fa24cf6e73deeed/src/cmd/compile/internal/ssa/gen/generic.rules#L586
			//
			//	(Slicemask (Const32 [0]))          => (Const32 [0])
			//
			Name:   "auxint",
			Source: "(Slicemask (Const32 [0]))          => (Const32 [0])\n",
			Expect: &ast.Rule{
				Match: &ast.SExpr{
					Op: [][]string{{"Slicemask"}},
					Args: []ast.Value{
						&ast.SExpr{
							Op:     [][]string{{"Const32"}},
							AuxInt: "0",
						},
					},
				},
				Result: &ast.SExpr{
					Op:     [][]string{{"Const32"}},
					AuxInt: "0",
				},
			},
		},

		{
			// Reference: https://github.com/golang/go/blob/02ce4118219dc51a14680a0c5fa24cf6e73deeed/src/cmd/compile/internal/ssa/gen/generic.rules#L109
			//
			//	(AddPtr <t> x (Const64 [c])) => (OffPtr <t> x [c])
			//
			Name:   "type",
			Source: "(AddPtr <t> x (Const64 [c])) => (OffPtr <t> x [c])\n",
			Expect: &ast.Rule{
				Match: &ast.SExpr{
					Op:   [][]string{{"AddPtr"}},
					Type: "t",
					Args: []ast.Value{
						ast.Variable("x"),
						&ast.SExpr{
							Op:     [][]string{{"Const64"}},
							AuxInt: "c",
						},
					},
				},
				Result: &ast.SExpr{
					Op:     [][]string{{"OffPtr"}},
					Type:   "t",
					AuxInt: "c",
					Args: []ast.Value{
						ast.Variable("x"),
					},
				},
			},
		},

		{
			// Reference: https://github.com/golang/go/blob/02ce4118219dc51a14680a0c5fa24cf6e73deeed/src/cmd/compile/internal/ssa/gen/generic.rules#L2016
			//
			//	(EqPtr  (Addr {x} _) (Addr {y} _)) => (ConstBool [x == y])
			//
			Name:   "aux",
			Source: "(EqPtr  (Addr {x} _) (Addr {y} _)) => (ConstBool [x == y])\n",
			Expect: &ast.Rule{
				Match: &ast.SExpr{
					Op: [][]string{{"EqPtr"}},
					Args: []ast.Value{
						&ast.SExpr{
							Op:   [][]string{{"Addr"}},
							Aux:  "x",
							Args: []ast.Value{ast.Variable("_")},
						},
						&ast.SExpr{
							Op:   [][]string{{"Addr"}},
							Aux:  "y",
							Args: []ast.Value{ast.Variable("_")},
						},
					},
				},
				Result: &ast.SExpr{
					Op:     [][]string{{"ConstBool"}},
					AuxInt: "x == y",
				},
			},
		},

		{
			// Reference: https://github.com/golang/go/blob/02ce4118219dc51a14680a0c5fa24cf6e73deeed/src/cmd/compile/internal/ssa/gen/generic.rules#L1962
			//
			//	(Mul(32|64)F x (Const(32|64)F [1])) => x
			//
			Name:   "alt2",
			Source: "(Mul(32|64)F x (Const(32|64)F [1])) => x\n",
			Expect: &ast.Rule{
				Match: &ast.SExpr{
					Op: [][]string{
						{"Mul"},
						{"32", "64"},
						{"F"},
					},
					Args: []ast.Value{
						ast.Variable("x"),
						&ast.SExpr{
							Op: [][]string{
								{"Const"},
								{"32", "64"},
								{"F"},
							},
							AuxInt: "1",
						},
					},
				},
				Result: ast.Variable("x"),
			},
		},

		{
			// Reference: https://github.com/golang/go/blob/02ce4118219dc51a14680a0c5fa24cf6e73deeed/src/cmd/compile/internal/ssa/gen/generic.rules#L326
			//
			//	(Not (Less(64|32|16|8) x y)) => (Leq(64|32|16|8) y x)
			//
			Name:   "alt4",
			Source: "(Not (Less(64|32|16|8) x y)) => (Leq(64|32|16|8) y x)\n",
			Expect: &ast.Rule{
				Match: &ast.SExpr{
					Op: [][]string{{"Not"}},
					Args: []ast.Value{
						&ast.SExpr{
							Op: [][]string{
								{"Less"},
								{"64", "32", "16", "8"},
							},
							Args: []ast.Value{
								ast.Variable("x"),
								ast.Variable("y"),
							},
						},
					},
				},
				Result: &ast.SExpr{
					Op: [][]string{
						{"Leq"},
						{"64", "32", "16", "8"},
					},
					Args: []ast.Value{
						ast.Variable("y"),
						ast.Variable("x"),
					},
				},
			},
		},

		{
			// Reference: https://github.com/golang/go/blob/02ce4118219dc51a14680a0c5fa24cf6e73deeed/src/cmd/compile/internal/ssa/gen/generic.rules#L100
			//
			//	(Neg32F (Const32F [c])) && c != 0 => (Const32F [-c])
			//
			Name:   "cond",
			Source: "(Neg32F (Const32F [c])) && c != 0 => (Const32F [-c])\n",
			Expect: &ast.Rule{
				Match: &ast.SExpr{
					Op: [][]string{{"Neg32F"}},
					Args: []ast.Value{
						&ast.SExpr{
							Op:     [][]string{{"Const32F"}},
							AuxInt: "c",
						},
					},
				},
				Conditions: []string{"c != 0"},
				Result: &ast.SExpr{
					Op:     [][]string{{"Const32F"}},
					AuxInt: "-c",
				},
			},
		},

		{
			// Reference: https://github.com/golang/go/blob/02ce4118219dc51a14680a0c5fa24cf6e73deeed/src/cmd/compile/internal/ssa/gen/generic.rules#L564
			//
			//	(Trunc64to8  (And64 (Const64 [y]) x)) && y&0xFF == 0xFF => (Trunc64to8 x)
			//
			Name:   "cond_with_ampersand",
			Source: "(Trunc64to8  (And64 (Const64 [y]) x)) && y&0xFF == 0xFF => (Trunc64to8 x)\n",
			Expect: &ast.Rule{
				Match: &ast.SExpr{
					Op: [][]string{{"Trunc64to8"}},
					Args: []ast.Value{
						&ast.SExpr{
							Op: [][]string{{"And64"}},
							Args: []ast.Value{
								&ast.SExpr{
									Op:     [][]string{{"Const64"}},
									AuxInt: "y",
								},
								ast.Variable("x"),
							},
						},
					},
				},
				Conditions: []string{"y&0xFF == 0xFF"},
				Result: &ast.SExpr{
					Op: [][]string{{"Trunc64to8"}},
					Args: []ast.Value{
						ast.Variable("x"),
					},
				},
			},
		},

		{
			// Reference: https://github.com/golang/go/blob/02ce4118219dc51a14680a0c5fa24cf6e73deeed/src/cmd/compile/internal/ssa/gen/generic.rules#L571
			//
			//	(ZeroExt8to64  (Trunc64to8  x:(Rsh64Ux64 _ (Const64 [s])))) && s >= 56 => x
			//
			Name:   "binding",
			Source: "(ZeroExt8to64  (Trunc64to8  x:(Rsh64Ux64 _ (Const64 [s])))) && s >= 56 => x\n",
			Expect: &ast.Rule{
				Match: &ast.SExpr{
					Op: [][]string{{"ZeroExt8to64"}},
					Args: []ast.Value{
						&ast.SExpr{
							Op: [][]string{{"Trunc64to8"}},
							Args: []ast.Value{
								&ast.SExpr{
									Binding: "x",
									Op:      [][]string{{"Rsh64Ux64"}},
									Args: []ast.Value{
										ast.Variable("_"),
										&ast.SExpr{
											Op:     [][]string{{"Const64"}},
											AuxInt: "s",
										},
									},
								},
							},
						},
					},
				},
				Conditions: []string{"s >= 56"},
				Result:     ast.Variable("x"),
			},
		},

		{
			// Reference: https://github.com/golang/go/blob/02ce4118219dc51a14680a0c5fa24cf6e73deeed/src/cmd/compile/internal/ssa/gen/386.rules#L580
			//
			//	(MOVBLSX x:(MOVBload [off] {sym} ptr mem)) && x.Uses == 1 && clobber(x) => @x.Block (MOVBLSXload <v.Type> [off] {sym} ptr mem)
			//
			Name:   "block",
			Source: "(MOVBLSX x:(MOVBload [off] {sym} ptr mem)) && x.Uses == 1 && clobber(x) => @x.Block (MOVBLSXload <v.Type> [off] {sym} ptr mem)\n",
			Expect: &ast.Rule{
				Match: &ast.SExpr{
					Op: [][]string{{"MOVBLSX"}},
					Args: []ast.Value{
						&ast.SExpr{
							Binding: "x",
							Op:      [][]string{{"MOVBload"}},
							AuxInt:  "off",
							Aux:     "sym",
							Args: []ast.Value{
								ast.Variable("ptr"),
								ast.Variable("mem"),
							},
						},
					},
				},
				Conditions: []string{"x.Uses == 1", "clobber(x)"},
				Block:      "x.Block",
				Result: &ast.SExpr{
					Op:     [][]string{{"MOVBLSXload"}},
					Type:   "v.Type",
					AuxInt: "off",
					Aux:    "sym",
					Args: []ast.Value{
						ast.Variable("ptr"),
						ast.Variable("mem"),
					},
				},
			},
		},

		{
			// Reference: https://github.com/golang/go/blob/02ce4118219dc51a14680a0c5fa24cf6e73deeed/src/cmd/compile/internal/ssa/gen/AMD64.rules#L7
			//
			//	(AddPtr ...) => (ADDQ ...)
			//
			Name:   "ellipsis",
			Source: "(AddPtr ...) => (ADDQ ...)",
			Expect: &ast.Rule{
				Match: &ast.SExpr{
					Op:       [][]string{{"AddPtr"}},
					Ellipsis: true,
				},
				Result: &ast.SExpr{
					Op:       [][]string{{"ADDQ"}},
					Ellipsis: true,
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			t.Logf("source =\n%s", c.Source)

			got, err := parse.String(c.Source)
			if err != nil {
				t.Fatal(err)
			}

			if c.Expect == nil {
				t.Fatal("expected rule not specified")
			}

			expect := &ast.File{Rules: []*ast.Rule{c.Expect}}

			if diff := cmp.Diff(expect, got); diff != "" {
				e, err := ast.Dump(expect)
				if err != nil {
					t.Fatal(err)
				}
				t.Logf("expect =\n%s", e)

				g, err := ast.Dump(got)
				if err != nil {
					t.Fatal(err)
				}
				t.Logf("got =\n%s", g)

				t.Logf("diff =\n%s", diff)

				t.FailNow()
			}
		})
	}
}
