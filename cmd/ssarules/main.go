package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/mmcloughlin/ssarules/ast"
	"github.com/mmcloughlin/ssarules/parse"
	"github.com/mmcloughlin/ssarules/pass"
	"github.com/mmcloughlin/ssarules/printer"
	"github.com/mmcloughlin/ssarules/rules"
)

func main() {
	log.SetPrefix("ssarules: ")
	log.SetFlags(0)

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	flag.Parse()
	filenames := flag.Args()

	// Parse filenames.
	var rs []*ast.Rule
	for _, filename := range filenames {
		f, err := parse.File(filename)
		if err != nil {
			return fmt.Errorf("parse: %w", err)
		}
		log.Printf("parsed %d rules from %s", len(f.Rules), filename)
		rs = append(rs, f.Rules...)
	}
	log.Printf("parsed %d rules total", len(rs))

	// Filter to simple rules.
	s := &ast.File{}
	for _, r := range rs {
		if !(len(r.Conditions) > 0 ||
			r.Block != "" ||
			rules.HasBinding(r) ||
			rules.HasType(r) ||
			rules.HasAux(r) ||
			rules.HasEllipsis(r) ||
			rules.HasTrailing(r)) {
			s.Rules = append(s.Rules, r)
		}
	}

	log.Printf("filter simple: %d rules", len(s.Rules))

	// Expand alternates.
	if err := pass.ExpandAlternates(s); err != nil {
		return err
	}

	log.Printf("expand alternates: %d rules", len(s.Rules))

	// Output.
	if err := printer.Print(s); err != nil {
		return err
	}

	return nil
}
