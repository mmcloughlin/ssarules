// Package printer implements printing of acc AST nodes.
package printer

import (
	"bytes"
	"fmt"
	goast "go/ast"
	goprinter "go/printer"
	"go/token"
	"io"
	"os"

	"github.com/mmcloughlin/ssarules/ast"
	"github.com/mmcloughlin/ssarules/internal/errutil"
)

// Bytes prints the AST and returns resulting bytes.
func Bytes(n ast.Node) ([]byte, error) {
	var buf bytes.Buffer
	if err := Fprint(&buf, n); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Print an AST node to standard out.
func Print(n ast.Node) error {
	return Fprint(os.Stdout, n)
}

// Fprint writes the AST node n to w.
func Fprint(w io.Writer, n ast.Node) error {
	p := newprinter(w)
	p.node(n)
	return p.err
}

type printer struct {
	w   io.Writer
	err error
}

func newprinter(w io.Writer) *printer {
	return &printer{w: w}
}

func (p *printer) node(node ast.Node) {
	switch n := node.(type) {
	case *ast.File:
		for _, r := range n.Rules {
			p.rule(r)
			p.printf("\n")
		}

	case *ast.Rule:
		p.rule(n)

	case ast.Value:
		p.value(n)

	case ast.Op:
		p.op(n)

	default:
		p.seterror(errutil.UnexpectedType(node))
	}
}

func (p *printer) rule(r *ast.Rule) {
	// Match
	p.value(r.Match)

	// Condition (optional)
	if r.Condition != nil {
		p.printf(" && ")
		p.goexpr(r.Condition)
	}

	// Deduction symbol
	p.printf(" => ")

	// Block (optional)
	if r.Block != "" {
		p.printf("@%s ", r.Block)
	}

	// Result
	p.value(r.Result)
}

func (p *printer) value(val ast.Value) {
	switch v := val.(type) {
	case *ast.SExpr:
		p.sexpr(v)
	case ast.Expr:
		p.goexpr(v.Expr)
	case ast.Variable:
		p.printf("%s", v)
	default:
		p.seterror(errutil.UnexpectedType(val))
	}
}

func (p *printer) sexpr(s *ast.SExpr) {
	// Binding (optional)
	if s.Binding != "" {
		p.printf("%s:", s.Binding)
	}

	// Open
	p.printf("(")

	// Op
	p.op(s.Op)

	// Type (optional)
	if s.Type != "" {
		p.printf(" <%s>", s.Type)
	}

	// AuxInt (optional)
	if s.AuxInt != "" {
		p.printf(" [%s]", s.AuxInt)
	}

	// Aux (optional)
	if s.Aux != "" {
		p.printf(" {%s}", s.Aux)
	}

	// Args
	for _, arg := range s.Args {
		p.printf(" ")
		p.value(arg)
	}

	// Ellipsis
	if s.Ellipsis {
		p.printf(" ...")
	}

	// Trailing
	if s.Trailing {
		p.printf(" ___")
	}

	// Close
	p.printf(")")
}

func (p *printer) op(op ast.Op) {
	switch o := op.(type) {
	case ast.Opcode:
		p.printf("%s", o)
	case ast.OpcodeParts:
		for _, part := range o {
			p.oppart(part)
		}
	default:
		p.seterror(errutil.UnexpectedType(op))
	}
}

func (p *printer) oppart(oppart ast.OpPart) {
	switch part := oppart.(type) {
	case ast.Opcode:
		p.printf("%s", part)
	case ast.OpcodeAlt:
		sep := '('
		for _, c := range part {
			p.printf("%c%s", sep, c)
			sep = '|'
		}
		p.printf(")")
	default:
		p.seterror(errutil.UnexpectedType(oppart))
	}
}

func (p *printer) goexpr(expr goast.Expr) {
	fset := token.NewFileSet()
	if err := goprinter.Fprint(p.w, fset, expr); err != nil {
		p.seterror(err)
	}
}

func (p *printer) printf(format string, args ...interface{}) {
	if p.err != nil {
		return
	}
	_, err := fmt.Fprintf(p.w, format, args...)
	p.seterror(err)
}

func (p *printer) seterror(err error) {
	if p.err == nil {
		p.err = err
	}
}
