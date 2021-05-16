package goexpr

import (
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/mmcloughlin/ssarules/internal/errutil"
)

func Parse(x string) (ast.Expr, error) {
	expr, err := parser.ParseExpr(x)
	if err != nil {
		return nil, err
	}

	if err := clearpos(expr); err != nil {
		return nil, err
	}

	return expr, nil
}

func clearpos(expr ast.Expr) error {
	var err error
	ast.Inspect(expr, func(node ast.Node) bool {
		if node == nil {
			return true
		}
		switch n := node.(type) {
		case *ast.BasicLit:
			n.ValuePos = token.NoPos
		case *ast.ParenExpr:
			n.Lparen = token.NoPos
			n.Rparen = token.NoPos
		case *ast.UnaryExpr:
			n.OpPos = token.NoPos
		case *ast.CallExpr:
			n.Lparen = token.NoPos
			n.Ellipsis = token.NoPos
			n.Rparen = token.NoPos
		case *ast.BinaryExpr:
			n.OpPos = token.NoPos
		case *ast.Ident:
			n.NamePos = token.NoPos
		case *ast.SelectorExpr:
			// pass
		default:
			err = errutil.UnexpectedType(node)
			return false
		}
		return true
	})
	return err
}
