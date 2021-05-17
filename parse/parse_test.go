package parse_test

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mmcloughlin/ssarules/ast"
	"github.com/mmcloughlin/ssarules/internal/test"
	"github.com/mmcloughlin/ssarules/parse"
	"github.com/mmcloughlin/ssarules/parse/internal/goexpr"
)

func TestFiles(t *testing.T) {
	test.Glob(t, "../internal/testdata/*.rules", func(t *testing.T, filename string) {
		// Parse file.
		f, err := parse.File(filename)
		if err != nil {
			t.Fatal(err)
		}

		// Check we parsed the right number of rules.
		expect := CountRules(t, filename)
		if len(f.Rules) != expect {
			t.Fatalf("parsed %d rules; expect %d", len(f.Rules), expect)
		}
	})
}

func CountRules(t *testing.T, filename string) int {
	t.Helper()

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	n := 0
	s := bufio.NewScanner(bytes.NewReader(b))
	for s.Scan() {
		line := s.Text()
		if i := strings.Index(line, "//"); i >= 0 {
			line = line[:i]
		}
		n += strings.Count(line, "=>")
	}

	return n
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
					Op: ast.Opcode("Slicemask"),
					Args: []ast.Value{
						&ast.SExpr{
							Op:     ast.Opcode("Const32"),
							AuxInt: BuildExpr(t, "0"),
						},
					},
				},
				Result: &ast.SExpr{
					Op:     ast.Opcode("Const32"),
					AuxInt: BuildExpr(t, "0"),
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
					Op:   ast.Opcode("AddPtr"),
					Type: "t",
					Args: []ast.Value{
						ast.Variable("x"),
						&ast.SExpr{
							Op:     ast.Opcode("Const64"),
							AuxInt: BuildExpr(t, "c"),
						},
					},
				},
				Result: &ast.SExpr{
					Op:     ast.Opcode("OffPtr"),
					Type:   "t",
					AuxInt: BuildExpr(t, "c"),
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
					Op: ast.Opcode("EqPtr"),
					Args: []ast.Value{
						&ast.SExpr{
							Op:   ast.Opcode("Addr"),
							Aux:  BuildExpr(t, "x"),
							Args: []ast.Value{ast.Placeholder},
						},
						&ast.SExpr{
							Op:   ast.Opcode("Addr"),
							Aux:  BuildExpr(t, "y"),
							Args: []ast.Value{ast.Placeholder},
						},
					},
				},
				Result: &ast.SExpr{
					Op:     ast.Opcode("ConstBool"),
					AuxInt: BuildExpr(t, "x == y"),
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
					Op: ast.OpcodeParts{
						ast.Opcode("Mul"),
						ast.OpcodeAlt{"32", "64"},
						ast.Opcode("F"),
					},
					Args: []ast.Value{
						ast.Variable("x"),
						&ast.SExpr{
							Op: ast.OpcodeParts{
								ast.Opcode("Const"),
								ast.OpcodeAlt{"32", "64"},
								ast.Opcode("F"),
							},
							AuxInt: BuildExpr(t, "1"),
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
					Op: ast.Opcode("Not"),
					Args: []ast.Value{
						&ast.SExpr{
							Op: ast.OpcodeParts{ast.Opcode("Less"), ast.OpcodeAlt{"64", "32", "16", "8"}},
							Args: []ast.Value{
								ast.Variable("x"),
								ast.Variable("y"),
							},
						},
					},
				},
				Result: &ast.SExpr{
					Op: ast.OpcodeParts{ast.Opcode("Leq"), ast.OpcodeAlt{"64", "32", "16", "8"}},
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
					Op: ast.Opcode("Neg32F"),
					Args: []ast.Value{
						&ast.SExpr{
							Op:     ast.Opcode("Const32F"),
							AuxInt: BuildExpr(t, "c"),
						},
					},
				},
				Condition: BuildExpr(t, "c != 0"),
				Result: &ast.SExpr{
					Op:     ast.Opcode("Const32F"),
					AuxInt: BuildExpr(t, "-c"),
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
					Op: ast.Opcode("Trunc64to8"),
					Args: []ast.Value{
						&ast.SExpr{
							Op: ast.Opcode("And64"),
							Args: []ast.Value{
								&ast.SExpr{
									Op:     ast.Opcode("Const64"),
									AuxInt: BuildExpr(t, "y"),
								},
								ast.Variable("x"),
							},
						},
					},
				},
				Condition: BuildExpr(t, "y&0xFF == 0xFF"),
				Result: &ast.SExpr{
					Op: ast.Opcode("Trunc64to8"),
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
					Op: ast.Opcode("ZeroExt8to64"),
					Args: []ast.Value{
						&ast.SExpr{
							Op: ast.Opcode("Trunc64to8"),
							Args: []ast.Value{
								&ast.SExpr{
									Binding: "x",
									Op:      ast.Opcode("Rsh64Ux64"),
									Args: []ast.Value{
										ast.Variable("_"),
										&ast.SExpr{
											Op:     ast.Opcode("Const64"),
											AuxInt: BuildExpr(t, "s"),
										},
									},
								},
							},
						},
					},
				},
				Condition: BuildExpr(t, "s >= 56"),
				Result:    ast.Variable("x"),
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
					Op: ast.Opcode("MOVBLSX"),
					Args: []ast.Value{
						&ast.SExpr{
							Binding: "x",
							Op:      ast.Opcode("MOVBload"),
							AuxInt:  BuildExpr(t, "off"),
							Aux:     BuildExpr(t, "sym"),
							Args: []ast.Value{
								ast.Variable("ptr"),
								ast.Variable("mem"),
							},
						},
					},
				},
				Condition: BuildExpr(t, "x.Uses == 1 && clobber(x)"),
				Block:     "x.Block",
				Result: &ast.SExpr{
					Op:     ast.Opcode("MOVBLSXload"),
					Type:   "v.Type",
					AuxInt: BuildExpr(t, "off"),
					Aux:    BuildExpr(t, "sym"),
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
					Op:       ast.Opcode("AddPtr"),
					Ellipsis: true,
				},
				Result: &ast.SExpr{
					Op:       ast.Opcode("ADDQ"),
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
					Op:     ast.Opcode("SelectN"),
					AuxInt: BuildExpr(t, "2"),
					Args: []ast.Value{
						&ast.SExpr{
							Op:       ast.Opcode("MakeResult"),
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
					Op: ast.Opcode("NE"),
					Args: []ast.Value{
						&ast.SExpr{
							Op: ast.Opcode("FlagLT_ULT"),
						},
						ast.Variable("yes"),
						ast.Variable("no"),
					},
				},
				Result: &ast.SExpr{
					Op: ast.Opcode("First"),
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
					Op:     ast.Opcode("InterLECall"),
					AuxInt: BuildExpr(t, "argsize"),
					Aux:    BuildExpr(t, "auxCall"),
					Args: []ast.Value{
						&ast.SExpr{
							Op: ast.Opcode("Load"),
							Args: []ast.Value{
								&ast.SExpr{
									Op:     ast.Opcode("OffPtr"),
									AuxInt: BuildExpr(t, "off"),
									Args: []ast.Value{
										&ast.SExpr{
											Op: ast.Opcode("ITab"),
											Args: []ast.Value{
												&ast.SExpr{
													Op: ast.Opcode("IMake"),
													Args: []ast.Value{
														&ast.SExpr{
															Op:  ast.Opcode("Addr"),
															Aux: BuildExpr(t, "itab"),
															Args: []ast.Value{
																&ast.SExpr{
																	Op: ast.Opcode("SB"),
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
				Condition: BuildExpr(t, "devirtLESym(v, auxCall, itab, off) !=\n\t\t\t\tnil"),
				Result:    BuildExpr(t, "devirtLECall(v, devirtLESym(v, auxCall, itab, off))"),
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
					Op: ast.Opcode("CondSelect"),
					Args: []ast.Value{
						ast.Variable("x"),
						ast.Variable("y"),
						ast.Variable("boolval"),
					},
				},
				Condition: BuildExpr(t, "flagArg(boolval) != nil"),
				Result: &ast.SExpr{
					Op:     ast.Opcode("CSEL"),
					AuxInt: BuildExpr(t, "boolval.Op"),
					Args: []ast.Value{
						ast.Variable("x"),
						ast.Variable("y"),
						BuildExpr(t, "flagArg(boolval)"),
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

func BuildExpr(t *testing.T, x string) *ast.Expr {
	expr, err := goexpr.Parse(x)
	if err != nil {
		t.Fatal(err)
	}
	return &ast.Expr{Expr: expr}
}
