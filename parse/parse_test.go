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

		{
			// Reference: https://github.com/golang/go/blob/c14ecaca8182314efd2ef7280feffc2242644887/src/cmd/compile/internal/ssa/gen/generic.rules#L1979
			//
			//	(SelectN [2] (MakeResult x y z ___)) => z
			//
			Name:   "trailing",
			Source: "(SelectN [2] (MakeResult x y z ___)) => z",
			Expect: &ast.Rule{
				Match: &ast.SExpr{
					Op:     [][]string{{"SelectN"}},
					AuxInt: "2",
					Args: []ast.Value{
						&ast.SExpr{
							Op:       [][]string{{"MakeResult"}},
							Args:     []ast.Value{ast.Variable("x"), ast.Variable("y"), ast.Variable("z")},
							Trailing: true,
						},
					},
				},
				Result: ast.Variable("z"),
			},
		},

		{
			// Reference: https://github.com/golang/go/blob/5203357ebacf9f41ca5e194d953c164049172e96/src/cmd/compile/internal/ssa/gen/386.rules#L773
			//
			//	(NE (FlagLT_ULT) yes no) => (First yes no)
			//
			Name:   "opcode_with_underscore",
			Source: "(NE (FlagLT_ULT) yes no) => (First yes no)",
			Expect: &ast.Rule{
				Match: &ast.SExpr{
					Op: [][]string{{"NE"}},
					Args: []ast.Value{
						&ast.SExpr{
							Op: [][]string{{"FlagLT_ULT"}},
						},
						ast.Variable("yes"),
						ast.Variable("no"),
					},
				},
				Result: &ast.SExpr{
					Op: [][]string{{"First"}},
					Args: []ast.Value{
						ast.Variable("yes"),
						ast.Variable("no"),
					},
				},
			},
		},

		{
			// Reference: https://github.com/golang/go/blob/02ce4118219dc51a14680a0c5fa24cf6e73deeed/src/cmd/compile/internal/ssa/gen/generic.rules#L2098-L2099
			//
			//	(InterLECall [argsize] {auxCall} (Load (OffPtr [off] (ITab (IMake (Addr {itab} (SB)) _))) _) ___) && devirtLESym(v, auxCall, itab, off) !=
			//	    nil => devirtLECall(v, devirtLESym(v, auxCall, itab, off))
			//
			Name: "expr_result",
			Source: `(InterLECall [argsize] {auxCall} (Load (OffPtr [off] (ITab (IMake (Addr {itab} (SB)) _))) _) ___) && devirtLESym(v, auxCall, itab, off) !=
				nil => devirtLECall(v, devirtLESym(v, auxCall, itab, off))`,
			Expect: &ast.Rule{
				Match: &ast.SExpr{
					Op:     [][]string{{"InterLECall"}},
					AuxInt: "argsize",
					Aux:    "auxCall",
					Args: []ast.Value{
						&ast.SExpr{
							Op: [][]string{{"Load"}},
							Args: []ast.Value{
								&ast.SExpr{
									Op:     [][]string{{"OffPtr"}},
									AuxInt: "off",
									Args: []ast.Value{
										&ast.SExpr{
											Op: [][]string{{"ITab"}},
											Args: []ast.Value{
												&ast.SExpr{
													Op: [][]string{{"IMake"}},
													Args: []ast.Value{
														&ast.SExpr{
															Op:  [][]string{{"Addr"}},
															Aux: "itab",
															Args: []ast.Value{
																&ast.SExpr{
																	Op: [][]string{{"SB"}},
																},
															},
														},
														ast.Placeholder,
													},
												},
											},
										},
									},
								},
								ast.Placeholder,
							},
						},
					},
					Trailing: true,
				},
				Conditions: []string{"devirtLESym(v, auxCall, itab, off) !=\n\t\t\t\tnil"},
				Result:     ast.Expr("devirtLECall(v, devirtLESym(v, auxCall, itab, off))"),
			},
		},

		{
			// Reference: https://github.com/golang/go/blob/5203357ebacf9f41ca5e194d953c164049172e96/src/cmd/compile/internal/ssa/gen/ARM64.rules#L320
			//
			//	(CondSelect x y boolval) && flagArg(boolval) != nil => (CSEL [boolval.Op] x y flagArg(boolval))
			//
			Name:   "expr_arg",
			Source: "(CondSelect x y boolval) && flagArg(boolval) != nil => (CSEL [boolval.Op] x y flagArg(boolval))\n",
			Expect: &ast.Rule{
				Match: &ast.SExpr{
					Op: [][]string{{"CondSelect"}},
					Args: []ast.Value{
						ast.Variable("x"),
						ast.Variable("y"),
						ast.Variable("boolval"),
					},
				},
				Conditions: []string{"flagArg(boolval) != nil"},
				Result: &ast.SExpr{
					Op:     [][]string{{"CSEL"}},
					AuxInt: "boolval.Op",
					Args: []ast.Value{
						ast.Variable("x"),
						ast.Variable("y"),
						ast.Expr("flagArg(boolval)"),
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got, err := parse.String(c.Source)
			if err != nil {
				t.Fatal(err)
			}

			if c.Expect == nil {
				t.Fatal("expected rule not specified")
			}

			expect := &ast.File{Rules: []*ast.Rule{c.Expect}}

			if diff := cmp.Diff(expect, got); diff != "" {
				t.Logf("source =\n%s", c.Source)

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
