// Package parse implements a parser for SSA rules.
package parse

import (
	"io"
	"strings"

	"github.com/mmcloughlin/ssarules/ast"
	"github.com/mmcloughlin/ssarules/parse/internal/parser"
)

//go:generate pigeon -o internal/parser/zparser.go rules.peg

// Reader parses the data from r using filename as information in
// error messages.
func Reader(filename string, r io.Reader) (*ast.File, error) {
	return cast(parser.ParseReader(filename, r))
}

// String parses s.
func String(s string) (*ast.File, error) {
	return Reader("string", strings.NewReader(s))
}

func cast(i interface{}, err error) (*ast.File, error) {
	if err != nil {
		return nil, err
	}
	return i.(*ast.File), nil
}
