package smt2

import (
	"fmt"
	gotypes "go/types"

	"github.com/mmcloughlin/ssarules/ast"
)

type value struct {
	Expr string
	Type gotypes.Type
}

type input struct {
	args []*value
}

type operation interface {
	Params(result gotypes.Type) ([]gotypes.Type, error)
	Evaluate(*input) (*value, error)
}

var operations = buildoperations()

func buildoperations() map[ast.Opcode]operation {
	operations := map[ast.Opcode]operation{}

	// Simple binary operations.
	for _, b := range []struct {
		name   ast.Opcode
		format string
	}{
		{"Add", "(bvadd %s %s)"},
		{"Sub", "(bvadd %s (bvneg %s))"},

		{"And", "(bvand %s %s)"},
		{"Or", "(bvor %s %s)"},
		{"Xor", "(bvxor %s %s)"},
	} {
		operations[b.name+"8"] = basicbinary(gotypes.Int8, b.format)
		operations[b.name+"16"] = basicbinary(gotypes.Int16, b.format)
		operations[b.name+"32"] = basicbinary(gotypes.Int32, b.format)
		operations[b.name+"64"] = basicbinary(gotypes.Int64, b.format)
	}

	return operations
}

type simple struct {
	result gotypes.Type
	params []gotypes.Type
	format string
}

func basicbinary(kind gotypes.BasicKind, format string) operation {
	t := gotypes.Typ[kind]
	return &simple{
		result: t,
		params: []gotypes.Type{t, t},
		format: format,
	}
}

func (s *simple) Params(result gotypes.Type) ([]gotypes.Type, error) {
	if result != nil && !gotypes.Identical(result, s.result) {
		return nil, fmt.Errorf("return type: expect %s got %s", s.result, result)
	}
	return s.params, nil
}

func (s *simple) Evaluate(in *input) (*value, error) {
	var args []interface{}
	for _, arg := range in.args {
		args = append(args, arg.Expr)
	}
	return &value{
		Expr: fmt.Sprintf(s.format, args...),
		Type: s.result,
	}, nil
}
