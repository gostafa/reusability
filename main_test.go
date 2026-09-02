// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"go/parser"
	"go/token"
	"os"
	"os/exec"
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
		"go/",
		"golang.org/x/tools",
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

var allowedProductionFiles = map[string]bool{
	"doc.go":    true,
	"consts.go": true,
	"funcs.go":  true,
	"types.go":  true,
	"vars.go":   true,
}

// isAllowedProductionFile reports whether name fits the declaration-kind layout.
// Extra types_*.go / *_types.go files are allowed so packages can stay under
// revive's max-public-structs limit without leaving the types kind.
func isAllowedProductionFile(name string) bool {
	if allowedProductionFiles[name] {
		return true
	}
	if strings.HasPrefix(name, "types_") && strings.HasSuffix(name, ".go") {
		return true
	}
	return strings.HasSuffix(name, "_types.go")
}

func isLayoutExcludedDir(dir string) bool {
	if strings.Contains(dir, string(filepath.Separator)+"testdata"+string(filepath.Separator)) {
		return true
	}
	return strings.HasSuffix(
		dir,
		string(filepath.Separator)+"tools"+string(filepath.Separator)+"layoutrefactor",
	)
}

func collectLayoutPackageDirs(t *testing.T) []string {
	t.Helper()

	out, err := exec.Command("go", "list", "-f", "{{.Dir}}", "./...").Output()
	if err != nil {
		t.Fatal(err)
	}

	var dirs []string
	for _, dir := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if dir == "" || isLayoutExcludedDir(dir) {
			continue
		}
		dirs = append(dirs, dir)
	}

	return dirs
}

func assertPackageFileLayout(t *testing.T, dir string) {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	var prodFiles, testFiles []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(entry.Name(), "_test.go") {
			testFiles = append(testFiles, entry.Name())
			continue
		}
		prodFiles = append(prodFiles, entry.Name())
	}

	if len(prodFiles) == 0 {
		return
	}

	hasDoc := false
	for _, name := range prodFiles {
		if name == "doc.go" {
			hasDoc = true
		}
		if !isAllowedProductionFile(name) {
			t.Errorf(
				"%s: disallowed production file %q (allowed: doc.go, consts.go, types.go, types_*.go, *_types.go, vars.go, funcs.go)",
				dir,
				name,
			)
		}
	}

	if !hasDoc {
		t.Errorf("%s: package has production code but no doc.go", dir)
	}

	if len(testFiles) > 1 {
		t.Errorf("%s: want at most one *_test.go file, got %d: %v", dir, len(testFiles), testFiles)
	}
}

// TestPackageFileLayout enforces the declaration-kind file layout in every
// production package: doc.go plus optional consts/types (including types_*.go
// / *_types.go splits)/vars/funcs, and at most one internal test file.
func TestPackageFileLayout(t *testing.T) {
	for _, dir := range collectLayoutPackageDirs(t) {
		t.Run(dir, func(t *testing.T) {
			assertPackageFileLayout(t, dir)
		})
	}
}
