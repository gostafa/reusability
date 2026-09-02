package reusability_test

import (
	"context"
	"math"
	"reflect"
	"sync"
	"testing"

	"github.com/gostafa/reusability/internal/shared/metrics"
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

func wantValue(t *testing.T, r reusability.MetricResult, want float64) {
	t.Helper()

	if !r.Applicable {
		t.Fatalf("reusability not applicable (%s), want %v", r.Reason, want)
	}

	if math.Abs(r.Value-want) > epsilon {
		t.Fatalf("reusability = %v, want %v", r.Value, want)
	}
}

func wantNotApplicable(t *testing.T, r reusability.MetricResult) {
	t.Helper()

	if r.Applicable {
		t.Fatalf("reusability applicable with value %v, want n/a", r.Value)
	}

	if r.Reason == "" {
		t.Fatalf("reusability n/a without reason")
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

	if order.Name != "Order" {
		t.Fatalf("order type name = %q, want Order", order.Name)
	}

	wantRI := 0.35*(1.0/3) + 0.25*0.5 + 0.25*0.6 + 0.15*(2.0/3)
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
			t.Fatalf("%s.%s reusability = %v, want reusability metric", loc.pkg, loc.typ, got.Reusability)
		}
	}
}

func TestAnalyzeTransitiveFieldUsage(t *testing.T) {
	direct := analyzeFixture(t, nil)
	counter := findType(t, findPackage(t, direct, "example.com/fixture/multifile"), "Counter")
	directRI := counter.Reusability

	transitive := analyzeFixture(t, func(cfg *reusability.Config) {
		cfg.FieldUsageMode = reusability.FieldUsageTransitive
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

	report = analyzeFixture(t, func(cfg *reusability.Config) { cfg.IncludeGenerated = true })
	gen = findPackage(t, report, "example.com/fixture/gen")
	machine := findType(t, gen, "Machine")
	if machine.Reusability.Name != metrics.MetricReusability {
		t.Fatalf("Machine reusability = %v, want reusability metric", machine.Reusability)
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
