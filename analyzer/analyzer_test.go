package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
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
	methods := maximum(0)

	r := newRunner(Settings{
		Directory: fixtureDir,
		Patterns:  []string{"./isolated"},
		Type:      &TypeSettings{Methods: &methods},
	}.withDefaults())

	r.load()
	if r.err != nil {
		t.Fatal(r.err)
	}

	got := r.byPkg["example.com/fixture/isolated"]
	if len(got) == 0 {
		t.Fatal("expected violations for isolated with methods max: 0")
	}

	foundMethods := false
	for _, v := range got {
		if v.Key == policydomain.KeyMethods && v.Type == "Value" {
			foundMethods = true
		}
	}

	if !foundMethods {
		t.Fatalf("expected methods violation on Value, got %#v", got)
	}
}

func maximum(value float64) LimitSettings {
	return LimitSettings{Max: &value}
}

func maximumFunc(value float64) FuncSettings {
	return FuncSettings{Max: &value}
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
		Package:    "example.com/p",
		Type:       "T",
		Key:        "methods",
		Value:      3,
		Comparator: policydomain.ComparatorMax,
		Threshold:  0,
	})

	want := "example.com/p.T (type): methods 3 exceeds max 0"
	if msg != want {
		t.Fatalf("formatViolation = %q, want %q", msg, want)
	}
}

func TestTypePosAndPackagePos(t *testing.T) {
	t.Parallel()

	src := `package p

const C = 1
var V int
func helper() {}
func Exported() {}
type Widget struct{}
func (Widget) Do() {}
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

	positionCases := map[string]string{
		policydomain.KeyFuncs:           "helper",
		policydomain.KeyExportedFuncs:   "Exported",
		policydomain.KeyUnexportedFuncs: "helper",
		policydomain.KeyVars:            "V",
		policydomain.KeyConsts:          "C",
	}
	for key, ident := range positionCases {
		want := identPos(t, file, ident)
		if pos := structuralPos(pass, key); pos != want {
			t.Fatalf("structuralPos(%q) = %v, want %v", key, pos, want)
		}
	}

	for _, tc := range []struct {
		receiver string
		name     string
	}{
		{"", "helper"},
		{"Widget", "Do"},
	} {
		want := identPos(t, file, tc.name)
		if pos := exactFuncPos(pass, tc.receiver, tc.name); pos != want {
			t.Fatalf("exactFuncPos(%q, %q) = %v, want %v",
				tc.receiver,
				tc.name,
				pos,
				want,
			)
		}
	}
}

func identPos(t *testing.T, file *ast.File, name string) token.Pos {
	t.Helper()

	var pos token.Pos
	ast.Inspect(file, func(n ast.Node) bool {
		if pos != token.NoPos {
			return false
		}

		ident, ok := n.(*ast.Ident)
		if !ok || ident.Name != name {
			return true
		}

		pos = ident.Pos()

		return false
	})

	if pos == token.NoPos {
		t.Fatalf("identifier %q not found", name)
	}

	return pos
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
}
