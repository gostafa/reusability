// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package reusability

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/gostafa/reusability/internal/shared/metrics"
)

const epsilon = 1e-12

// The default-config report is loaded once and shared read-only across the
// many default-config tests — package loading dominates test time, so this
// avoids re-running the analyzer for every case. Config-varying tests (mutate
// != nil) still load fresh.
var (
	defaultOnce   sync.Once
	defaultReport Report
	defaultErr    error
)

func analyzeFixture(t *testing.T, mutate func(*Config)) Report {
	t.Helper()

	if mutate == nil {
		defaultOnce.Do(func() {
			defaultReport, defaultErr = Analyze(
				context.Background(), &Config{Directory: "testdata/fixture"},
			)
		})

		if defaultErr != nil {
			t.Fatal(defaultErr)
		}

		return defaultReport
	}

	cfg := Config{Directory: "testdata/fixture"}
	mutate(&cfg)

	report, err := Analyze(context.Background(), &cfg)
	if err != nil {
		t.Fatal(err)
	}

	return report
}

func findPackage(t *testing.T, report Report, path string) PackageReport {
	t.Helper()

	for _, pkg := range report.Packages {
		if pkg.Path == path {
			return pkg
		}
	}

	t.Fatalf("package %s not in report", path)

	return PackageReport{}
}

func findType(t *testing.T, pkg PackageReport, name string) TypeReport {
	t.Helper()

	for _, typ := range pkg.Types {
		if typ.Name == name {
			return typ
		}
	}

	t.Fatalf("type %s not in package %s", name, pkg.Path)

	return TypeReport{}
}

func wantValue(t *testing.T, r MetricResult, want float64) {
	t.Helper()

	if !r.Applicable {
		t.Fatalf("reusability not applicable (%s), want %v", r.Reason, want)
	}

	if math.Abs(r.Value-want) > epsilon {
		t.Fatalf("reusability = %v, want %v", r.Value, want)
	}
}

func TestAnalyzeFixtureOrdering(t *testing.T) {
	report := analyzeFixture(t, nil)

	wantOrder := []string{
		"example.com/fixture/embedding",
		"example.com/fixture/gen",
		"example.com/fixture/generics",
		"example.com/fixture/isolated",
		"example.com/fixture/multifile",
		"example.com/fixture/orders",
		"example.com/fixture/store",
	}
	if len(report.Packages) != len(wantOrder) {
		t.Fatalf("got %d packages", len(report.Packages))
	}

	for i, path := range wantOrder {
		if report.Packages[i].Path != path {
			t.Fatalf("packages[%d] = %s, want %s", i, report.Packages[i].Path, path)
		}
	}

	if report.SchemaVersion != SchemaVersion ||
		report.Tool.Name != ToolName() {
		t.Fatalf("report header = %+v", report)
	}
}

func TestAnalyzeOrderType(t *testing.T) {
	report := analyzeFixture(t, nil)
	order := findType(t, findPackage(t, report, "example.com/fixture/orders"), "Order")

	if order.Name != "Order" {
		t.Fatalf("order type name = %q, want Order", order.Name)
	}

	wantRI := 0.30*(1.0/3) + 0.35*0.5 + 0.20*0.6 + 0.15*(2.0/3)
	wantValue(t, order.Reusability, wantRI)

	if order.Reusability.Name != metrics.MetricReusability {
		t.Fatalf("unexpected metric name %q", order.Reusability.Name)
	}
}

func TestAnalyzePackageCouplingFacts(t *testing.T) {
	report := analyzeFixture(t, nil)

	store := findPackage(t, report, "example.com/fixture/store")
	storeType := findType(t, store, "Store")
	if storeType.Reusability.Name != metrics.MetricReusability {
		t.Fatalf("Store reusability = %v, want reusability metric", storeType.Reusability)
	}
}

func TestAnalyzeGenericsAndEmbedding(t *testing.T) {
	report := analyzeFixture(t, nil)

	for _, loc := range []struct {
		pkg, typ string
	}{
		{"example.com/fixture/generics", "Pair"},
		{"example.com/fixture/embedding", "Wrapper"},
		{"example.com/fixture/embedding", "Base"},
	} {
		got := findType(t, findPackage(t, report, loc.pkg), loc.typ)
		if got.Reusability.Name != metrics.MetricReusability {
			t.Fatalf(
				"%s.%s reusability = %v, want reusability metric",
				loc.pkg,
				loc.typ,
				got.Reusability,
			)
		}
	}
}

func TestAnalyzeTransitiveFieldUsage(t *testing.T) {
	direct := analyzeFixture(t, nil)
	counter := findType(t, findPackage(t, direct, "example.com/fixture/multifile"), "Counter")
	directRI := counter.Reusability

	transitive := analyzeFixture(t, func(cfg *Config) {
		cfg.FieldUsageMode = FieldUsageTransitive
	})
	counter = findType(t, findPackage(t, transitive, "example.com/fixture/multifile"), "Counter")
	transitiveRI := counter.Reusability
	if !directRI.Applicable || !transitiveRI.Applicable {
		t.Fatalf("reusability n/a: direct %+v transitive %+v", directRI, transitiveRI)
	}

	if math.Abs(directRI.Value-transitiveRI.Value) < epsilon {
		t.Fatalf("transitive field usage should change reusability; both %v", directRI.Value)
	}
}

func TestAnalyzeGeneratedFiles(t *testing.T) {
	report := analyzeFixture(t, nil)

	gen := findPackage(t, report, "example.com/fixture/gen")
	if len(gen.Types) != 0 {
		t.Fatalf("generated types analyzed by default: %v", gen.Types)
	}

	report = analyzeFixture(t, func(cfg *Config) { cfg.IncludeGenerated = true })
	gen = findPackage(t, report, "example.com/fixture/gen")
	machine := findType(t, gen, "Machine")
	if machine.Reusability.Name != metrics.MetricReusability {
		t.Fatalf("Machine reusability = %v, want reusability metric", machine.Reusability)
	}
}

func TestAnalyzeDeterminism(t *testing.T) {
	first := analyzeFixture(t, func(cfg *Config) { cfg.Workers = 1 })

	second := analyzeFixture(t, func(cfg *Config) { cfg.Workers = 8 })
	if !reflect.DeepEqual(first, second) {
		t.Fatal("reports differ across worker counts")
	}

	third := analyzeFixture(t, func(cfg *Config) { cfg.Workers = 8 })
	if !reflect.DeepEqual(second, third) {
		t.Fatal("repeated runs differ")
	}
}

func TestAnalyzeInvalidConfig(t *testing.T) {
	ctx := context.Background()
	base := Config{Directory: "testdata/fixture"}

	bad := base

	bad.DependencyScope = "galaxy"
	if _, err := Analyze(ctx, &bad); err == nil {
		t.Fatal("invalid scope accepted")
	}

	bad = base

	bad.ReusabilityWeights = Weights{Cohesion: -1, Coupling: 2}
	if _, err := Analyze(ctx, &bad); err == nil {
		t.Fatal("negative weight accepted")
	}
}

func TestAnalyzeCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := Analyze(
		ctx, &Config{Directory: "testdata/fixture"},
	); err == nil {
		t.Fatal("cancelled context accepted")
	}
}

func TestWithDefaults(t *testing.T) {
	cfg := configWithDefaults(&Config{})
	if len(cfg.Patterns) != 1 || cfg.Patterns[0] != defaultPackagePattern {
		t.Fatalf("patterns = %v", cfg.Patterns)
	}

	if cfg.DependencyScope != DependencyScopeModule {
		t.Fatalf("scope = %q", cfg.DependencyScope)
	}

	if cfg.FieldUsageMode != FieldUsageDirect {
		t.Fatalf("field usage = %q", cfg.FieldUsageMode)
	}

	if cfg.ReusabilityWeights != metrics.DefaultReusabilityWeights() {
		t.Fatalf("weights = %+v", cfg.ReusabilityWeights)
	}
}

func TestValidate(t *testing.T) {
	valid := configWithDefaults(&Config{})
	err := validateConfig(&valid)
	if err != nil {
		t.Fatal(err)
	}

	bad := valid

	bad.DependencyScope = "galaxy"
	err = validateConfig(&bad)
	if err == nil {
		t.Fatal("invalid scope accepted")
	}

	bad = valid

	bad.FieldUsageMode = "psychic"
	err = validateConfig(&bad)
	if err == nil {
		t.Fatal("invalid field usage accepted")
	}

	bad = valid

	bad.Patterns = []string{""}
	err = validateConfig(&bad)
	if err == nil {
		t.Fatal("empty pattern accepted")
	}

	bad = valid

	bad.ReusabilityWeights = Weights{Cohesion: -0.5, Coupling: 1}
	err = validateConfig(&bad)
	if err == nil {
		t.Fatal("negative weight accepted")
	}
}

func TestAllMetrics(t *testing.T) {
	got := AllMetrics()
	if len(got) != 1 || got[0] != MetricReusability {
		t.Fatalf("AllMetrics() = %v, want [%s]", got, MetricReusability)
	}

	if def := DefaultMetrics(); len(def) != 1 || def[0] != MetricReusability {
		t.Fatalf("DefaultMetrics() = %v", def)
	}
}

func TestRepositoryMeetsUniformReusabilityPolicy(t *testing.T) {
	root, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}

	report, err := Analyze(context.Background(), &Config{
		Directory: root,
		Patterns:  []string{"./..."},
	})
	if err != nil {
		t.Fatalf("analyze repository: %v", err)
	}

	var items []string
	for i := range report.Packages {
		pkg := &report.Packages[i]
		for j := range pkg.Types {
			typ := &pkg.Types[j]
			if typ.Reusability.Applicable && typ.Reusability.Value < 0.60 {
				items = append(items, pkg.Path+"."+typ.Name)
			}
		}
	}
	if len(items) == 0 {
		return
	}

	t.Fatalf("uniform reusability policy violations: %s", strings.Join(items, ", "))
}

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
	"consts.go": true,
	"funcs.go":  true,
	"types.go":  true,
	"vars.go":   true,
}

// isAllowedProductionFile reports whether name fits the declaration-kind layout.
func isAllowedProductionFile(name string) bool {
	if allowedProductionFiles[name] {
		return true
	}
	return strings.HasPrefix(name, "types_") && strings.HasSuffix(name, ".go")
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
	fset := token.NewFileSet()
	packageName := ""
	for _, name := range prodFiles {
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		packageName = file.Name.Name
		if name == "types.go" {
			hasDoc = file.Doc != nil
		}
		if !isAllowedProductionFile(name) {
			t.Errorf(
				"%s: disallowed production file %q (allowed: consts.go, types.go, types_*.go, vars.go, funcs.go)",
				dir,
				name,
			)
		}
		for _, declaration := range file.Decls {
			want := "funcs.go"
			if declaration, ok := declaration.(*ast.GenDecl); ok {
				switch declaration.Tok {
				case token.IMPORT:
					continue
				case token.TYPE:
					want = "types.go"
					if strings.HasPrefix(name, "types_") {
						want = name
					}
				case token.CONST:
					want = "consts.go"
				case token.VAR:
					want = "vars.go"
				}
			}
			if name != want {
				t.Errorf("%s: declaration in %s belongs in %s", dir, name, want)
			}
		}
	}

	if !hasDoc {
		t.Errorf("%s: package has production code but no types.go", dir)
	}

	for _, name := range testFiles {
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.PackageClauseOnly)
		if err != nil {
			t.Fatal(err)
		}
		valid := name == packageName+"_test.go" || name == packageName+"_bench_test.go"
		if file.Name.Name == packageName+"_test" {
			valid = name == packageName+"_integration_test.go"
		} else if file.Name.Name != packageName {
			valid = false
		}
		if !valid {
			t.Errorf("%s: test filename %s does not match package %s", dir, name, file.Name.Name)
		}
	}
}

// TestPackageFileLayout enforces the declaration-kind file layout in every
// production package, including declaration placement and test package names.
func TestPackageFileLayout(t *testing.T) {
	for _, dir := range collectLayoutPackageDirs(t) {
		t.Run(dir, func(t *testing.T) {
			assertPackageFileLayout(t, dir)
		})
	}
}
