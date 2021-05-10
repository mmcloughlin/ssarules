package ast

import (
	"github.com/mmcloughlin/ssarules/internal/errutil"
)

// A Visitor's Visit method is invoked for each node encountered by Walk.  If
// the result visitor w is not nil, Walk visits each of the children of node
// with the visitor w.
type Visitor interface {
	Visit(node Node) (w Visitor)
}

func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	switch n := node.(type) {
	case *File:
		for _, r := range n.Rules {
			Walk(v, r)
		}

	case *Rule:
		Walk(v, n.Match)
		Walk(v, n.Result)

	case *SExpr:
		Walk(v, n.Op)
		for _, arg := range n.Args {
			Walk(v, arg)
		}

	case OpcodeParts:
		for _, part := range n {
			Walk(v, part)
		}

	default:
		panic(errutil.UnexpectedType(node))
	}
}

// Inspect traverses an AST in depth-first order: It starts by calling f(node);
// node must not be nil. If f returns true, Inspect invokes f recursively for
// each of the non-nil children of node.
func Inspect(node Node, f func(Node) bool) {
	Walk(inspector(f), node)
}

type inspector func(Node) bool

func (f inspector) Visit(node Node) Visitor {
	if f(node) {
		return f
	}
	return nil
}
