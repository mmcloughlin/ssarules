package rules

import "github.com/mmcloughlin/ssarules/ast"

func HasBinding(r *ast.Rule) bool {
	return ast.Contains(r, func(n ast.Node) bool {
		s, ok := n.(*ast.SExpr)
		return ok && (s.Binding != "")
	})
}

func HasType(r *ast.Rule) bool {
	return ast.Contains(r, func(n ast.Node) bool {
		s, ok := n.(*ast.SExpr)
		return ok && s.Type != ""
	})
}

func HasAux(r *ast.Rule) bool {
	return ast.Contains(r, func(n ast.Node) bool {
		s, ok := n.(*ast.SExpr)
		return ok && s.Aux != nil
	})
}

func HasEllipsis(r *ast.Rule) bool {
	return ast.Contains(r, func(n ast.Node) bool {
		s, ok := n.(*ast.SExpr)
		return ok && s.Ellipsis
	})
}

func HasTrailing(r *ast.Rule) bool {
	return ast.Contains(r, func(n ast.Node) bool {
		s, ok := n.(*ast.SExpr)
		return ok && s.Trailing
	})
}
