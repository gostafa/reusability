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

func TestRunnerRunReportsPackageAndTypeViolations(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "p.go", "package p\n\ntype Widget struct{}\n", 0)
	if err != nil {
		t.Fatal(err)
	}

	r := &runner{byPkg: map[string][]policydomain.Violation{
		"example.com/p": {
			{
				Package:    "example.com/p",
				Key:        policydomain.KeyTypes,
				Value:      2,
				Comparator: policydomain.ComparatorMax,
				Threshold:  1,
			},
			{
				Package:    "example.com/p",
				Type:       "Widget",
				Key:        policydomain.KeyMethods,
				Value:      0.5,
				Comparator: policydomain.ComparatorMin,
				Threshold:  1.25,
			},
		},
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
	if len(diagnostics) != 2 {
		t.Fatalf("diagnostics = %#v, want two", diagnostics)
	}
	if diagnostics[0].Pos != file.Package {
		t.Errorf("package diagnostic position = %v, want %v", diagnostics[0].Pos, file.Package)
	}
	if !strings.Contains(diagnostics[1].Message, "is below min 1.25") {
		t.Errorf("type diagnostic = %q", diagnostics[1].Message)
	}

	sentinel := errors.New("cached load error")
	failing := &runner{err: sentinel}
	failing.once.Do(func() {})
	if _, err := failing.run(pass); !errors.Is(err, sentinel) {
		t.Fatalf("run error = %v, want sentinel", err)
	}
}

func TestRunnerLoadErrors(t *testing.T) {
	r := newRunner(Settings{Type: &TypeSettings{Metrics: map[string]LimitSettings{
		"lcom": maximum(1),
	}}}.withDefaults())
	r.load()
	if r.err == nil || !strings.Contains(r.err.Error(), "reusability policy") {
		t.Fatalf("policy load error = %v", r.err)
	}

	r = newRunner(Settings{
		Directory: filepath.Join(t.TempDir(), "missing"),
		Patterns:  []string{"./..."},
	}.withDefaults())
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

	defaults, err := (Settings{Directory: dir}).policy()
	if err != nil {
		t.Fatal(err)
	}
	if !defaults.Package.Types.HasMax || defaults.Package.Types.Max != 12 {
		t.Fatalf("default types limit = %+v, want max 12", defaults.Package.Types)
	}

	types := maximum(3)
	funcs := maximumFunc(6)
	funcLines := maximum(80)
	funcCyclomatic := maximum(10)
	vars := maximum(4)
	consts := maximum(5)
	inline, err := (Settings{Package: &PackageSettings{
		Types:  &types,
		Funcs:  &funcs,
		Vars:   &vars,
		Consts: &consts,
	}, Funcs: &FuncSettings{
		Lines:      &funcLines,
		Cyclomatic: &funcCyclomatic,
	}}).policy()
	if err != nil {
		t.Fatal(err)
	}
	if !inline.Package.Types.HasMax || inline.Package.Types.Max != 3 {
		t.Fatalf("inline types limit = %+v", inline.Package.Types)
	}
	if !inline.Package.Funcs.Count.HasMax || inline.Package.Funcs.Count.Max != 6 {
		t.Fatalf("inline funcs limit = %+v", inline.Package.Funcs)
	}
	if !inline.Funcs.Lines.HasMax || inline.Funcs.Lines.Max != 80 {
		t.Fatalf("inline funcs lines limit = %+v", inline.Funcs)
	}
	if !inline.Funcs.Cyclomatic.HasMax || inline.Funcs.Cyclomatic.Max != 10 {
		t.Fatalf("inline funcs cyclomatic limit = %+v", inline.Funcs)
	}
	if !inline.Package.Vars.HasMax || inline.Package.Vars.Max != 4 {
		t.Fatalf("inline vars limit = %+v", inline.Package.Vars)
	}
	if !inline.Package.Consts.HasMax || inline.Package.Consts.Max != 5 {
		t.Fatalf("inline consts limit = %+v", inline.Package.Consts)
	}
	if inline.Type.Methods.HasMax {
		t.Fatalf(
			"inline policy unexpectedly merged default methods limit: %+v",
			inline.Type.Methods,
		)
	}
}

func TestEmptyPackagePosition(t *testing.T) {
	if pos := packagePos(&analysis.Pass{}); pos != token.NoPos {
		t.Fatalf("packagePos(empty) = %v, want NoPos", pos)
	}
}
