package smt2

import (
	"bytes"
	"errors"
	"fmt"
	gotypes "go/types"
	"io"
	"strings"

	"github.com/mmcloughlin/ssarules/ast"
	"github.com/mmcloughlin/ssarules/internal/errutil"
	"github.com/mmcloughlin/ssarules/printer"
)

func Generate(r *ast.Rule) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	g := &generator{
		w:    buf,
		ops:  operations,
		vars: map[ast.Variable]*value{},
	}
	if err := g.generate(r); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type generator struct {
	w    io.Writer
	ops  map[ast.Opcode]operation
	vars map[ast.Variable]*value
	err  error
}

func (g *generator) generate(r *ast.Rule) error {
	// Header
	g.printf("(set-info :smt-lib-version 2.6)\n")
	g.printf("(set-logic QF_BV)\n")

	// Include rule in comment
	rule, err := printer.Bytes(r)
	if err != nil {
		return err
	}
	g.printf("; %s\n", strings.ReplaceAll(string(rule), "\n", "\n; "))

	// Match
	match, err := g.eval(r.Match, nil)
	if err != nil {
		return err
	}

	// Condition
	if r.Condition != nil {
		return errutil.NotSupported("rule conditions")
	}

	// Block
	if r.Block != "" {
		return errutil.NotSupported("rule blocks")
	}

	// Result
	result, err := g.eval(r.Result, match.Type)
	if err != nil {
		return err
	}

	// Assert equivalence
	g.printf("(assert (not (= %s %s)))\n", match.Expr, result.Expr)
	g.printf("(check-sat)\n")

	// Exit
	g.printf("(exit)\n")

	return g.err
}

func (g *generator) eval(val ast.Value, t gotypes.Type) (*value, error) {
	switch v := val.(type) {
	case ast.Variable:
		return g.variable(v, t)
	case *ast.SExpr:
		return g.sexpr(v, t)
	// case *ast.Expr:
	// 	p.goexpr(v.Expr)
	default:
		return nil, errutil.UnexpectedType(val)
	}
}

func (g *generator) variable(v ast.Variable, t gotypes.Type) (*value, error) {
	// Has variable been declared already?
	if val, ok := g.vars[v]; ok {
		if t != nil && !gotypes.Identical(val.Type, t) {
			return nil, fmt.Errorf("inconsistent types for variable %s", v)
		}
		return val, nil
	}

	// Need to know the type in order to declare.
	if t == nil {
		return nil, fmt.Errorf("declare %s: cannot deduce type", v)
	}

	s, err := typesort(t)
	if err != nil {
		return nil, fmt.Errorf("declare %s: %w", v, err)
	}

	g.printf("(declare-const %s %s)\n", v, s)
	g.vars[v] = &value{
		Expr: string(v),
		Type: t,
	}

	return g.vars[v], nil
}

func typesort(t gotypes.Type) (string, error) {
	b, ok := t.(*gotypes.Basic)
	if !ok {
		return "", errutil.NotSupported("type %s", t)
	}

	switch b.Kind() {
	case gotypes.Bool:
		return "Bool", nil
	case gotypes.Int8, gotypes.Uint8:
		return "(_ BitVec 8)", nil
	case gotypes.Int16, gotypes.Uint16:
		return "(_ BitVec 16)", nil
	case gotypes.Int32, gotypes.Uint32:
		return "(_ BitVec 32)", nil
	case gotypes.Int64, gotypes.Uint64:
		return "(_ BitVec 64)", nil

	// 	Int
	// 	Uint
	// 	Uintptr
	// 	Float32
	// 	Float64
	// 	Complex64
	// 	Complex128
	// 	String
	// 	UnsafePointer

	default:
		return "", errutil.NotSupported("type %s", t)
	}
}

func (g *generator) sexpr(s *ast.SExpr, t gotypes.Type) (*value, error) {
	// Unsupported features
	switch {
	case s.Binding != "":
		return nil, errutil.NotSupported("sexpr binding")
	case s.Type != "":
		return nil, errutil.NotSupported("sexpr type field")
	case s.AuxInt != nil:
		return nil, errutil.NotSupported("sexpr auxint field")
	case s.Aux != nil:
		return nil, errutil.NotSupported("sexpr aux field")
	case s.Ellipsis:
		return nil, errutil.NotSupported("sexpr ellipsis")
	case s.Trailing:
		return nil, errutil.NotSupported("sexpr trailing")
	}

	// Lookup operation.
	opcode, ok := s.Op.(ast.Opcode)
	if !ok {
		return nil, errors.New("unexpanded opcode")
	}

	op, ok := g.ops[opcode]
	if !ok {
		return nil, errutil.NotSupported("opcode %s", opcode)
	}

	// Evaluate arguments.
	params, err := op.Params(t)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opcode, err)
	}

	if len(params) != len(s.Args) {
		return nil, fmt.Errorf("%s: expect %d arguments got %d", opcode, len(params), len(s.Args))
	}

	in := &input{}
	for i, arg := range s.Args {
		v, err := g.eval(arg, params[i])
		if err != nil {
			return nil, fmt.Errorf("%s: evaluate argument %d: %w", opcode, i, err)
		}
		in.args = append(in.args, v)
	}

	// Evalute this operation.
	v, err := op.Evaluate(in)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opcode, err)
	}

	return v, nil
}

func (g *generator) printf(format string, args ...interface{}) {
	if g.err != nil {
		return
	}
	_, err := fmt.Fprintf(g.w, format, args...)
	g.seterror(err)
}

func (g *generator) seterror(err error) {
	if g.err == nil && err != nil {
		g.err = err
	}
}
