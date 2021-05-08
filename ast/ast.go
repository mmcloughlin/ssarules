package ast

import (
	goast "go/ast"
	"io"
	"os"
)

type File struct {
	Rules []*Rule
}

type Rule struct {
	Match  Value
	Result Value
}

type Value interface{}

type SExpr struct {
	Op     Op
	Type   Type
	AuxInt AuxInt
	Aux    Aux
	Args   []Value
}

type Op [][]string

type Type string

type AuxInt string

type Aux string

type Variable string

// Print an AST node to standard out.
func Print(n interface{}) error {
	return Fprint(os.Stdout, n)
}

// Fprint writes the AST node n to w.
func Fprint(w io.Writer, n interface{}) error {
	return goast.Fprint(w, nil, n, goast.NotNilFilter)
}
