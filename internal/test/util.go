package test

import (
	"path/filepath"
	"strings"
	"testing"
)

// Glob runs a test on all the files matching a glob pattern.
func Glob(t *testing.T, pattern string, f func(*testing.T, string)) {
	t.Helper()

	filenames, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatal(err)
	}

	for _, filename := range filenames {
		base := filepath.Base(filename)
		ext := filepath.Ext(base)
		name := strings.TrimSuffix(base, ext)
		t.Run(name, func(t *testing.T) {
			f(t, filename)
		})
	}
}
