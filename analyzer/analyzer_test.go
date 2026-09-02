// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package analyzer

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"

	policydomain "github.com/gostafa/reusability/internal/features/policy/domain"
	"golang.org/x/tools/go/analysis"
)

func ptr(value float64) *float64 {
	return &value
}

func TestRunnerRunReportsPackageAndTypeViolations(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "p.go", "package p\n\ntype Widget struct{}\n", 0)
	if err != nil {
		t.Fatal(err)
	}

	r := &runner{byPkg: map[string][]policydomain.Violation{
		"example.com/p": {{
			Package:   "example.com/p",
			Type:      "Widget",
			Value:     0.5,
			Threshold: 0.8,
			Rule:      "**",
		}},
	}}
	r.once.Do(func() {})

	var diagnostics []analysis.Diagnostic
	pass := &analysis.Pass{
		Fset:   fset,
		Files:  []*ast.File{file},
		Pkg:    types.NewPackage("example.com/p", "p"),
		Report: func(d analysis.Diagnostic) { diagnostics = append(diagnostics, d) },
	}

	if _, err := r.run(pass); err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	if !strings.Contains(diagnostics[0].Message, "is below min 0.8") {
		t.Errorf("type diagnostic = %q", diagnostics[0].Message)
	}

	sentinel := errors.New("cached load error")
	failing := &runner{err: sentinel}
	failing.once.Do(func() {})
	if _, err := failing.run(pass); !errors.Is(err, sentinel) {
		t.Fatalf("run error = %v, want sentinel", err)
	}
}

func TestRunnerLoadErrors(t *testing.T) {
	min := 2.0
	settings := Settings{Rules: []RuleSettings{{Pattern: "**", Min: &min}}}
	s := settingsWithDefaults(&settings)
	r := newRunner(&s)
	r.load()
	if r.err == nil || !strings.Contains(r.err.Error(), "reusability policy") {
		t.Fatalf("policy load error = %v", r.err)
	}

	settings = Settings{
		Directory: filepath.Join(t.TempDir(), "missing"),
		Patterns:  []string{defaultPackagePattern},
	}
	s = settingsWithDefaults(&settings)
	r = newRunner(&s)
	r.load()
	if r.err == nil || !strings.Contains(r.err.Error(), "reusability analyze") {
		t.Fatalf("analysis load error = %v", r.err)
	}
}

func TestInlinePolicyDefaultsAndIgnoresModularityFile(t *testing.T) {
	dir := t.TempDir()
	config := filepath.Join(dir, ".modularity.yml")
	if err := os.WriteFile(
		config,
		[]byte("version: 1\npackage:\n  types: 3\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	defaultsSettings := Settings{Directory: dir}
	defaults, err := settingsRules(&defaultsSettings)
	if err != nil {
		t.Fatal(err)
	}
	if len(defaults) != 1 || defaults[0].Pattern != "**" || defaults[0].Min != 0.7 {
		t.Fatalf("default rules = %+v", defaults)
	}

	min := 0.6
	inlineSettings := Settings{Rules: []RuleSettings{
		{Pattern: "**/internal/**", Min: ptr(0.8)},
		{Pattern: "**", Min: &min},
	}}
	inline, err := settingsRules(&inlineSettings)
	if err != nil {
		t.Fatal(err)
	}
	if len(inline) != 2 || inline[0].Min != 0.8 || inline[1].Min != 0.6 {
		t.Fatalf("inline rules = %+v", inline)
	}
}

func TestEmptyPackagePosition(t *testing.T) {
	if pos := packagePos(&analysis.Pass{}); pos != token.NoPos {
		t.Fatalf("packagePos(empty) = %v, want NoPos", pos)
	}
}
