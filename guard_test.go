package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestForbiddenIdentifier enforces the project constraint that a certain
// legacy identifier never appears anywhere in the tree. The needle is
// assembled at runtime so this file cannot violate the rule itself.
func TestForbiddenIdentifier(t *testing.T) {
	needle := "go" + "metrics"

	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if name := d.Name(); name == ".git" || name == ".qodo" {
				return filepath.SkipDir
			}

			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if strings.Contains(strings.ToLower(string(content)), needle) {
			t.Errorf("%s contains the forbidden identifier", path)
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestDomainPurity enforces the architecture invariants: the public metrics
// package and every feature domain package import no compiler, CLI, JSON,
// filesystem, or logging package.
func TestDomainPurity(t *testing.T) {
	forbiddenPrefixes := []string{
		"go/",                // compiler stack (go/ast, go/types, go/token, …)
		"golang.org/x/tools", // compiler stack
		"encoding/json",
		"encoding/csv",
		"os",
		"io/fs",
		"io/ioutil",
		"path/filepath",
		"flag",
		"log",
	}

	var dirs []string

	dirs = append(dirs, "internal/shared/metrics")

	features, err := os.ReadDir("internal/features")
	if err != nil {
		t.Fatal(err)
	}

	for _, feature := range features {
		if !feature.IsDir() {
			continue
		}

		domainDir := filepath.Join("internal/features", feature.Name(), "domain")
		if _, err := os.Stat(domainDir); err == nil {
			dirs = append(dirs, domainDir)
		}
	}

	fset := token.NewFileSet()

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
				continue
			}

			if strings.HasSuffix(entry.Name(), "_test.go") {
				continue
			}

			path := filepath.Join(dir, entry.Name())

			file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatal(err)
			}

			for _, imp := range file.Imports {
				importPath := strings.Trim(imp.Path.Value, `"`)
				for _, prefix := range forbiddenPrefixes {
					if importPath == prefix || strings.HasPrefix(importPath, prefix+"/") {
						t.Errorf("%s imports %q, forbidden in pure packages", path, importPath)
					}
				}
			}
		}
	}
}
