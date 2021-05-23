package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/mmcloughlin/ssarules/ast"
	"github.com/mmcloughlin/ssarules/parse"
	"github.com/mmcloughlin/ssarules/pass"
	"github.com/mmcloughlin/ssarules/printer"
	"github.com/mmcloughlin/ssarules/rules"
	"github.com/mmcloughlin/ssarules/smt2"
)

func main() {
	log.SetPrefix("ssarules: ")
	log.SetFlags(0)

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

var outdir = flag.String("outdir", "", "output directory")

func run() error {
	flag.Parse()
	filenames := flag.Args()

	// Load and filter rules.
	f, err := load(filenames)
	if err != nil {
		return err
	}

	// Generate SMT.
	for i, r := range f.Rules {
		rule, err := printer.Bytes(r)
		if err != nil {
			return err
		}
		log.Printf("rule: %s", string(rule))

		b, err := smt2.Generate(r)
		if err != nil {
			log.Printf("error: %s", err)
			continue
		}

		filename := filepath.Join(*outdir, fmt.Sprintf("%04d.smt2", i))
		if err := ioutil.WriteFile(filename, b, 0644); err != nil {
			return err
		}

		log.Printf("smt2: written to %s", filename)
	}

	return nil
}

func load(filenames []string) (*ast.File, error) {
	// Parse filenames.
	var rs []*ast.Rule
	for _, filename := range filenames {
		f, err := parse.File(filename)
		if err != nil {
			return nil, fmt.Errorf("parse: %w", err)
		}
		log.Printf("parsed %d rules from %s", len(f.Rules), filename)
		rs = append(rs, f.Rules...)
	}
	log.Printf("parsed %d rules total", len(rs))

	// Filter to simple rules.
	s := &ast.File{}
	for _, r := range rs {
		if !(r.Condition != nil ||
			r.Block != "" ||
			rules.HasBinding(r) ||
			rules.HasType(r) ||
			rules.HasAux(r) ||
			rules.HasAuxInt(r) ||
			rules.HasEllipsis(r) ||
			rules.HasTrailing(r)) {
			s.Rules = append(s.Rules, r)
		}
	}

	log.Printf("filter simple: %d rules", len(s.Rules))

	// Expand alternates.
	if err := pass.ExpandAlternates(s); err != nil {
		return nil, err
	}

	log.Printf("expand alternates: %d rules", len(s.Rules))

	return s, nil
}
