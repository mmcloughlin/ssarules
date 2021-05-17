package pass

import (
	"fmt"

	"github.com/mmcloughlin/ssarules/ast"
	"github.com/mmcloughlin/ssarules/internal/errutil"
)

func ExpandAlternates(f *ast.File) error {
	rs, err := expandrules(f.Rules)
	if err != nil {
		return err
	}
	f.Rules = rs
	return nil
}

func expandrules(rs []*ast.Rule) ([]*ast.Rule, error) {
	var exps []*ast.Rule
	for _, r := range rs {
		exp, err := expandrule(r)
		if err != nil {
			return nil, err
		}
		exps = append(exps, exp...)
	}
	return exps, nil
}

func expandrule(r *ast.Rule) ([]*ast.Rule, error) {
	ms, err := expandvalue(r.Match)
	if err != nil {
		return nil, err
	}

	rs, err := expandvalue(r.Result)
	if err != nil {
		return nil, err
	}

	n, err := resolvecounts(len(ms), len(rs))
	if err != nil {
		return nil, err
	}

	if n == 1 {
		return []*ast.Rule{r}, nil
	}

	rules := make([]*ast.Rule, 0, n)
	for i := 0; i < n; i++ {
		rules = append(rules, &ast.Rule{
			Match:     ms[idx(i, len(ms))],
			Condition: r.Condition,
			Block:     r.Block,
			Result:    rs[idx(i, len(rs))],
		})
	}

	return rules, nil
}

func expandvalue(v ast.Value) ([]ast.Value, error) {
	switch val := v.(type) {
	case *ast.SExpr:
		return expandsexpr(val)
	case *ast.Expr, ast.Variable:
		return []ast.Value{val}, nil
	default:
		return nil, errutil.UnexpectedType(v)
	}
}

func expandsexpr(s *ast.SExpr) ([]ast.Value, error) {
	// Expand opcodes.
	opcodes, err := expandop(s.Op)
	if err != nil {
		return nil, err
	}

	n := len(opcodes)

	// Expand arguments.
	var expargs [][]ast.Value
	for _, arg := range s.Args {
		exparg, err := expandvalue(arg)
		if err != nil {
			return nil, err
		}

		n, err = resolvecounts(n, len(exparg))
		if err != nil {
			return nil, err
		}

		expargs = append(expargs, exparg)
	}

	if n == 1 {
		return []ast.Value{s}, nil
	}

	// Generate alternate SExpr values.
	vs := make([]ast.Value, 0, n)
	for i := 0; i < n; i++ {
		// Start with a clone of the original.
		v := new(ast.SExpr)
		*v = *s

		// Replace opcode.
		v.Op = opcodes[idx(i, len(opcodes))]

		// Replace args.
		var args []ast.Value
		for _, alts := range expargs {
			args = append(args, alts[idx(i, len(alts))])
		}
		v.Args = args

		vs = append(vs, v)
	}

	return vs, nil
}

func expandop(op ast.Op) ([]ast.Opcode, error) {
	switch o := op.(type) {
	case ast.Opcode:
		return []ast.Opcode{o}, nil
	case ast.OpcodeParts:
		return expandopcodeparts(o)
	default:
		return nil, errutil.UnexpectedType(op)
	}
}

func expandopcodeparts(parts ast.OpcodeParts) ([]ast.Opcode, error) {
	if len(parts) == 0 {
		return nil, errutil.AssertionFailure("empty opcode parts")
	}

	// Determine alternate count.
	n := 1
	for _, part := range parts {
		if alt, ok := part.(ast.OpcodeAlt); ok {
			var err error
			n, err = resolvecounts(n, len(alt))
			if err != nil {
				return nil, err
			}
		}
	}

	if n == 1 {
		return nil, errutil.AssertionFailure("multi-part opcode should represent multiple alternatives")
	}

	// Construct alternates.
	alts := make([]ast.Opcode, n)
	for i := 0; i < n; i++ {
		for _, part := range parts {
			switch p := part.(type) {
			case ast.Opcode:
				alts[i] += p
			case ast.OpcodeAlt:
				alts[i] += p[i]
			default:
				return nil, errutil.UnexpectedType(part)
			}
		}
	}

	return alts, nil
}

func resolvecounts(m, n int) (int, error) {
	switch {
	case m <= 0, n <= 0:
		return 0, errutil.AssertionFailure("expect positive counts")
	case m > 1 && n > 1 && m != n:
		return 0, fmt.Errorf("incompatible alternative counts %d and %d", m, n)
	case m > n:
		return m, nil
	default:
		return n, nil
	}
}

func idx(i, n int) int {
	if n == 1 {
		return 0
	}
	return i
}
