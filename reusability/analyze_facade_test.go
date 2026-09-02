package reusability_test

import (
	"context"
	"math"
	"reflect"
	"sync"
	"testing"

	"github.com/gostafa/reusability/reusability"
)

const epsilon = 1e-12

// The default-config report is loaded once and shared read-only across the
// many default-config tests — package loading dominates test time, so this
// avoids re-running the analyzer for every case. Config-varying tests (mutate
// != nil) still load fresh.
var (
	defaultOnce   sync.Once
	defaultReport reusability.Report
	defaultErr    error
)

func analyzeFixture(t *testing.T, mutate func(*reusability.Config)) reusability.Report {
	t.Helper()

	if mutate == nil {
		defaultOnce.Do(func() {
			defaultReport, defaultErr = reusability.Analyze(
				context.Background(), reusability.Config{Directory: "../testdata/fixture"},
			)
		})

		if defaultErr != nil {
			t.Fatal(defaultErr)
		}

		return defaultReport
	}

	cfg := reusability.Config{Directory: "../testdata/fixture"}
	mutate(&cfg)

	report, err := reusability.Analyze(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	return report
}

func findPackage(t *testing.T, report reusability.Report, path string) reusability.PackageReport {
	t.Helper()

	for _, pkg := range report.Packages {
		if pkg.Path == path {
			return pkg
		}
	}

	t.Fatalf("package %s not in report", path)

	return reusability.PackageReport{}
}

func findType(t *testing.T, pkg reusability.PackageReport, name string) reusability.TypeReport {
	t.Helper()

	for _, typ := range pkg.Types {
		if typ.Name == name {
			return typ
		}
	}

	t.Fatalf("type %s not in package %s", name, pkg.Path)

	return reusability.TypeReport{}
}

func metric(
	t *testing.T,
	results []reusability.MetricResult,
	name string,
) reusability.MetricResult {
	t.Helper()

	for _, r := range results {
		if r.Name == name {
			return r
		}
	}

	t.Fatalf("metric %s not present in %v", name, results)

	return reusability.MetricResult{}
}

func wantValue(t *testing.T, results []reusability.MetricResult, name string, want float64) {
	t.Helper()

	r := metric(t, results, name)
	if !r.Applicable {
		t.Fatalf("%s not applicable (%s), want %v", name, r.Reason, want)
	}

	if math.Abs(r.Value-want) > epsilon {
		t.Fatalf("%s = %v, want %v", name, r.Value, want)
	}
}

func wantNotApplicable(t *testing.T, results []reusability.MetricResult, name string) {
	t.Helper()

	r := metric(t, results, name)
	if r.Applicable {
		t.Fatalf("%s applicable with value %v, want n/a", name, r.Value)
	}

	if r.Reason == "" {
		t.Fatalf("%s n/a without reason", name)
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

	if report.SchemaVersion != reusability.SchemaVersion ||
		report.Tool.Name != reusability.ToolName {
		t.Fatalf("report header = %+v", report)
	}
}

func TestAnalyzeOrderType(t *testing.T) {
	report := analyzeFixture(t, nil)
	order := findType(t, findPackage(t, report, "example.com/fixture/orders"), "Order")

	if order.Kind != "struct" || !order.Exported || order.Position.File != "orders/orders.go" {
		t.Fatalf("order type details = %+v", order)
	}
	if len(order.FieldDetails) != 3 || order.FieldDetails[0].Name != "ID" {
		t.Fatalf("order field details = %+v", order.FieldDetails)
	}
	if len(order.MethodDetails) != 3 {
		t.Fatalf("order method details = %+v", order.MethodDetails)
	}
	grandTotal := findFunctionReport(t, order.MethodDetails, "GrandTotal")
	if grandTotal.Receiver != "Order" || grandTotal.Cyclomatic != 3 || grandTotal.Lines != 6 {
		t.Fatalf("GrandTotal details = %+v", grandTotal)
	}

	wantRI := 0.35*(1.0/3) + 0.25*0.5 + 0.25*0.6 + 0.15*(2.0/3)
	wantValue(t, order.Metrics, "reusability", wantRI)

	// Supporting metrics are computed internally but never displayed.
	for _, r := range order.Metrics {
		if r.Name != "reusability" {
			t.Fatalf("unexpected displayed metric %q", r.Name)
		}
	}
}

func findFunctionReport(t *testing.T, functions []reusability.FunctionReport, name string) reusability.FunctionReport {
	t.Helper()

	for _, fn := range functions {
		if fn.Name == name {
			return fn
		}
	}

	t.Fatalf("function %s not found in %+v", name, functions)

	return reusability.FunctionReport{}
}

func TestAnalyzePackageStructuralFacts(t *testing.T) {
	report := analyzeFixture(t, nil)

	orders := findPackage(t, report, "example.com/fixture/orders")
	if orders.Vars != 0 || orders.Consts != 0 {
		t.Fatalf("orders counts = vars %d consts %d, want 0 and 0", orders.Vars, orders.Consts)
	}

	store := findPackage(t, report, "example.com/fixture/store")
	storeType := findType(t, store, "Store")
	if len(storeType.Metrics) != 1 || storeType.Metrics[0].Name != "reusability" {
		t.Fatalf("Store metrics = %v, want reusability only", storeType.Metrics)
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
		if len(got.Metrics) != 1 || got.Metrics[0].Name != "reusability" {
			t.Fatalf("%s.%s metrics = %v, want reusability only", loc.pkg, loc.typ, got.Metrics)
		}
	}
}

func TestAnalyzeTransitiveFieldUsage(t *testing.T) {
	direct := analyzeFixture(t, nil)
	counter := findType(t, findPackage(t, direct, "example.com/fixture/multifile"), "Counter")
	directRI := metric(t, counter.Metrics, "reusability")

	transitive := analyzeFixture(t, func(cfg *reusability.Config) {
		cfg.FieldUsageMode = reusability.FieldUsageTransitive
	})
	counter = findType(t, findPackage(t, transitive, "example.com/fixture/multifile"), "Counter")
	transitiveRI := metric(t, counter.Metrics, "reusability")
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

	report = analyzeFixture(t, func(cfg *reusability.Config) { cfg.IncludeGenerated = true })
	gen = findPackage(t, report, "example.com/fixture/gen")
	machine := findType(t, gen, "Machine")
	if len(machine.Metrics) != 1 || machine.Metrics[0].Name != "reusability" {
		t.Fatalf("Machine metrics = %v, want reusability", machine.Metrics)
	}
}

func TestAnalyzeDeterminism(t *testing.T) {
	first := analyzeFixture(t, func(cfg *reusability.Config) { cfg.Workers = 1 })

	second := analyzeFixture(t, func(cfg *reusability.Config) { cfg.Workers = 8 })
	if !reflect.DeepEqual(first, second) {
		t.Fatal("reports differ across worker counts")
	}

	third := analyzeFixture(t, func(cfg *reusability.Config) { cfg.Workers = 8 })
	if !reflect.DeepEqual(second, third) {
		t.Fatal("repeated runs differ")
	}
}

func TestAnalyzeInvalidConfig(t *testing.T) {
	ctx := context.Background()
	base := reusability.Config{Directory: "../testdata/fixture"}

	bad := base

	bad.DependencyScope = "galaxy"
	if _, err := reusability.Analyze(ctx, bad); err == nil {
		t.Fatal("invalid scope accepted")
	}

	bad = base

	bad.ReusabilityWeights = reusability.ReusabilityWeights{Cohesion: -1, Coupling: 2}
	if _, err := reusability.Analyze(ctx, bad); err == nil {
		t.Fatal("negative weight accepted")
	}
}

func TestAnalyzeCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := reusability.Analyze(
		ctx,
		reusability.Config{Directory: "../testdata/fixture"},
	); err == nil {
		t.Fatal("cancelled context accepted")
	}
}
