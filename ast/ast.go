package ast

import (
	"bytes"
	goast "go/ast"
	"io"
	"os"
)

type Node interface {
	node() // sealed
}

type File struct {
	Rules []*Rule
}

func (*File) node() {}

type Rule struct {
	Match      Value
	Conditions []string
	Block      string
	Result     Value
}

func (*Rule) node() {}

type Value interface {
	Node
	value() // sealed
}

type SExpr struct {
	Binding  Variable
	Op       Op
	Type     Type
	AuxInt   AuxInt
	Aux      Aux
	Args     []Value
	Ellipsis bool
	Trailing bool
}

func (*SExpr) node()  {}
func (*SExpr) value() {}

type Op interface {
	Node
	op() // sealed
}

type Opcode string

func (Opcode) node()   {}
func (Opcode) oppart() {}

func (Opcode) op() {}

type OpcodeParts []OpPart

func (OpcodeParts) node() {}
func (OpcodeParts) op()   {}

type OpPart interface {
	Node
	oppart() // sealed
}

type OpcodeAlt []Opcode

func (OpcodeAlt) node()   {}
func (OpcodeAlt) oppart() {}

type Type string

type AuxInt string

type Aux string

type Expr string

func (Expr) node()  {}
func (Expr) value() {}

type Variable string

func (Variable) node()  {}
func (Variable) value() {}

var Placeholder = Variable("_")

// Print an AST node to standard out.
func Print(n Node) error {
	return Fprint(os.Stdout, n)
}

// Dump produces a string representation of the AST node.
func Dump(n Node) (string, error) {
	buf := bytes.NewBuffer(nil)
	if err := Fprint(buf, n); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Fprint writes the AST node n to w.
func Fprint(w io.Writer, n Node) error {
	return goast.Fprint(w, nil, n, goast.NotNilFilter)
}
