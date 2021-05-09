package ast

import (
	"bytes"
	goast "go/ast"
	"io"
	"os"
)

type File struct {
	Rules []*Rule
}

type Rule struct {
	Match      Value
	Conditions []string
	Block      string
	Result     Value
}

type Value interface{}

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

type Op interface {
	op() // sealed
}

func (Opcode) op()      {}
func (OpcodeParts) op() {}

type Opcode string

type OpcodeParts []OpPart

type OpPart interface {
	oppart() // sealed
}

func (Opcode) oppart()    {}
func (OpcodeAlt) oppart() {}

type OpcodeAlt []Opcode

type Type string

type AuxInt string

type Aux string

type Expr string

type Variable string

var Placeholder = Variable("_")

// Print an AST node to standard out.
func Print(n interface{}) error {
	return Fprint(os.Stdout, n)
}

// Dump produces a string representation of
func Dump(n interface{}) (string, error) {
	buf := bytes.NewBuffer(nil)
	if err := Fprint(buf, n); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Fprint writes the AST node n to w.
func Fprint(w io.Writer, n interface{}) error {
	return goast.Fprint(w, nil, n, goast.NotNilFilter)
}
