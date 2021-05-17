package goexpr

import (
	goast "go/ast"
	goparser "go/parser"
	gotoken "go/token"

	"github.com/mmcloughlin/ssarules/internal/errutil"
)

func Parse(x string) (goast.Expr, error) {
	expr, err := goparser.ParseExpr(x)
	if err != nil {
		return nil, err
	}

	if err := clearpos(expr); err != nil {
		return nil, err
	}

	return expr, nil
}

func clearpos(expr goast.Expr) error {
	var err error
	goast.Inspect(expr, func(node goast.Node) bool {
		if node == nil {
			return true
		}
		switch n := node.(type) {
		case *goast.BasicLit:
			n.ValuePos = gotoken.NoPos
		case *goast.ParenExpr:
			n.Lparen = gotoken.NoPos
			n.Rparen = gotoken.NoPos
		case *goast.UnaryExpr:
			n.OpPos = gotoken.NoPos
		case *goast.CallExpr:
			n.Lparen = gotoken.NoPos
			n.Ellipsis = gotoken.NoPos
			n.Rparen = gotoken.NoPos
		case *goast.BinaryExpr:
			n.OpPos = gotoken.NoPos
		case *goast.Ident:
			n.NamePos = gotoken.NoPos
		case *goast.StarExpr:
			n.Star = gotoken.NoPos
		case *goast.SelectorExpr:
			// pass
		default:
			err = errutil.UnexpectedType(node)
			return false
		}
		return true
	})
	return err
}
