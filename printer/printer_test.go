package printer_test

import (
	"bytes"
	"flag"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mmcloughlin/ssarules/internal/test"
	"github.com/mmcloughlin/ssarules/parse"
	"github.com/mmcloughlin/ssarules/printer"
)

var golden = flag.Bool("golden", false, "write golden testdata files")

func TestGolden(t *testing.T) {
	test.Glob(t, "../internal/testdata/*.rules", func(t *testing.T, filename string) {
		// Parse file.
		f, err := parse.File(filename)
		if err != nil {
			t.Fatal(err)
		}

		// Print to buffer.
		got, err := printer.Bytes(f)
		if err != nil {
			t.Fatal(err)
		}

		// Compare to golden file.
		name := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
		goldenfile := filepath.Join("testdata", name+".golden")

		if *golden {
			if err := ioutil.WriteFile(goldenfile, got, 0644); err != nil {
				t.Fatal(err)
			}
		}

		expect, err := ioutil.ReadFile(goldenfile)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(expect, got); diff != "" {
			t.Fatalf("diff =\n%s", diff)
		}
	})
}

func TestRoundtrip(t *testing.T) {
	test.Glob(t, "../internal/testdata/*.rules", func(t *testing.T, filename string) {
		// Parse file.
		f, err := parse.File(filename)
		if err != nil {
			t.Fatal(err)
		}

		// Print it.
		buf := bytes.NewBuffer(nil)
		if err := printer.Fprint(buf, f); err != nil {
			t.Fatal(err)
		}

		// Parse it again.
		f2, err := parse.Reader("buffer", buf)
		if err != nil {
			t.Fatal(err)
		}

		// Verify roundtrip.
		if diff := cmp.Diff(f, f2); diff != "" {
			t.Logf("diff = \n%s", diff)
			t.Fatal("roundtrip failure")
		}
	})
}
