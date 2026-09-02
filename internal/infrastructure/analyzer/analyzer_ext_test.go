package analyzer_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gostafa/reusability/internal/features/projectanalysis/ports/inbound"
	"github.com/gostafa/reusability/internal/infrastructure/analyzer"
	"github.com/gostafa/reusability/internal/shared/metrics"
)

// Black-box: the wired analyzer runs the real pipeline over the fixture module
// end to end (compiler load → facts → metrics).
func TestAnalyzeFixture(t *testing.T) {
	t.Parallel()

	result, err := analyzer.NewAnalyzer().Analyze(context.Background(), inbound.Options{
		Directory:       filepath.Join("..", "..", "..", "testdata", "fixture"),
		Patterns:        []string{"./..."},
		DependencyScope: "module",
		Weights:         metrics.DefaultReusabilityWeights(),
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.ModulePath != "example.com/fixture" {
		t.Fatalf("module = %q", result.ModulePath)
	}

	if len(result.Packages) < 7 {
		t.Fatalf("packages = %d, want >= 7", len(result.Packages))
	}
	// Packages come back sorted by import path.
	for i := 1; i < len(result.Packages); i++ {
		if result.Packages[i-1].Path > result.Packages[i].Path {
			t.Fatalf(
				"packages not sorted: %s before %s",
				result.Packages[i-1].Path,
				result.Packages[i].Path,
			)
		}
	}
}
