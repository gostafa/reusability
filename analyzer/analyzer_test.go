package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	policydomain "github.com/gostafa/reusability/internal/features/policy/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
	"golang.org/x/tools/go/analysis"
)

func TestNewRejectsInvalidSettings(t *testing.T) {
	t.Parallel()

	_, err := New(Settings{DependencyScope: "nope"})
	if err == nil {
		t.Fatal("expected error for invalid dependency-scope")
	}

	_, err = New(Settings{FieldUsage: "nope"})
	if err == nil {
		t.Fatal("expected error for invalid field-usage")
	}

	_, err = New(Settings{ReusabilityWeights: &ReusabilityWeightSettings{
		Cohesion: ptr(-1),
	}})
	if err == nil {
		t.Fatal("expected error for invalid reusability weight")
	}

	_, err = New(Settings{ReusabilityWeights: &ReusabilityWeightSettings{
		Cohesion:      ptr(0),
		Coupling:      ptr(0),
		Testability:   ptr(0),
		Documentation: ptr(0),
	}})
	if err == nil {
		t.Fatal("expected error for all-zero reusability weights")
	}

	min := 1.5
	_, err = New(Settings{Rules: []RuleSettings{{Pattern: "**", Min: &min}}})
	if err == nil {
		t.Fatal("expected error for invalid rule min")
	}
}

func TestNewAcceptsDefaults(t *testing.T) {
	t.Parallel()

	a, err := New(Settings{})
	if err != nil {
		t.Fatal(err)
	}

	if a.Name != Name {
		t.Fatalf("Name = %q, want %q", a.Name, Name)
	}
}

func TestRunnerLoadGroupsViolations(t *testing.T) {
	fixtureDir := filepath.Join(repoRoot(t), "testdata", "fixture")
	min := 0.99

	r := newRunner(Settings{
		Directory: fixtureDir,
		Patterns:  []string{"./isolated"},
		Rules:     []RuleSettings{{Pattern: "**", Min: &min}},
	}.withDefaults())

	r.load()
	if r.err != nil {
		t.Fatal(r.err)
	}

	got := r.byPkg["example.com/fixture/isolated"]
	if len(got) == 0 {
		t.Fatal("expected reusability violations for isolated with min 0.99")
	}

	found := false
	for _, v := range got {
		if v.Type == "Value" && v.Rule == "**" {
			found = true
		}
	}

	if !found {
		t.Fatalf("expected reusability violation on Value, got %#v", got)
	}
}

func ptr(value float64) *float64 {
	return &value
}

func TestReusabilityWeightsConfig(t *testing.T) {
	t.Parallel()

	settings := Settings{
		ReusabilityWeights: &ReusabilityWeightSettings{
			Cohesion:      ptr(0.1),
			Coupling:      ptr(0.2),
			Testability:   ptr(0.3),
			Documentation: ptr(0.4),
		},
	}.withDefaults()

	got := settings.toConfig().ReusabilityWeights
	want := metrics.ReusabilityWeights{
		Cohesion:      0.1,
		Coupling:      0.2,
		Testability:   0.3,
		Documentation: 0.4,
	}
	if got != want {
		t.Fatalf("weights = %+v, want %+v", got, want)
	}

	partial := Settings{
		ReusabilityWeights: &ReusabilityWeightSettings{
			Coupling: ptr(0),
		},
	}.withDefaults()

	got = partial.toConfig().ReusabilityWeights
	want = metrics.DefaultReusabilityWeights()
	want.Coupling = 0
	if got != want {
		t.Fatalf("partial weights = %+v, want %+v", got, want)
	}
}

func TestFormatViolation(t *testing.T) {
	t.Parallel()

	msg := formatViolation(policydomain.Violation{
		Package:   "example.com/p",
		Type:      "T",
		Value:     0.55,
		Threshold: 0.8,
		Rule:      "**/internal/**",
	})

	want := "example.com/p.T (type): reusability 0.55 is below min 0.80 (rule **/internal/**)"
	if msg != want {
		t.Fatalf("formatViolation = %q, want %q", msg, want)
	}
}

func TestTypePosAndPackagePos(t *testing.T) {
	t.Parallel()

	src := `package p

type Widget struct{}
`
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "p.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	pass := &analysis.Pass{Files: []*ast.File{file}, Fset: fset}

	if pos := typePos(pass, "Widget"); pos == token.NoPos {
		t.Fatal("typePos(Widget) = NoPos")
	}

	if pos := typePos(pass, "Missing"); pos != file.Package {
		t.Fatalf("typePos(Missing) = %v, want package clause", pos)
	}

	if pos := packagePos(pass); pos != file.Package {
		t.Fatalf("packagePos = %v, want %v", pos, file.Package)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
}

func TestViolationPosUsesType(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "p.go", "package p\n\ntype Widget struct{}\n", 0)
	if err != nil {
		t.Fatal(err)
	}

	pass := &analysis.Pass{Files: []*ast.File{file}, Fset: fset}
	pos := violationPos(pass, policydomain.Violation{Type: "Widget"})
	if pos == token.NoPos {
		t.Fatal("violationPos = NoPos")
	}
}

func TestRunnerRunReportsTypeViolations(t *testing.T) {
	t.Parallel()

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
		t.Errorf("diagnostic = %q", diagnostics[0].Message)
	}
}
