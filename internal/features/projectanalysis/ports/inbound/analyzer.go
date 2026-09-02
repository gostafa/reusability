package inbound

import (
	"context"
	"fmt"

	"github.com/gostafa/reusability/internal/shared/metrics"
)

// Options is a fully validated, defaults-applied analysis request.
type Options struct {
	// Directory is the working directory package loading runs from.
	Directory string
	// Patterns are the package patterns to analyze (e.g. "./...").
	Patterns []string
	// IncludeTests also analyzes test files and test packages.
	IncludeTests bool
	// IncludeGenerated also analyzes generated files.
	IncludeGenerated bool
	// BuildTags are extra build tags applied while loading.
	BuildTags []string
	// Workers bounds package-level concurrency; 0 selects a default.
	Workers int
	// DependencyScope is "project", "module", or "all".
	DependencyScope string
	// FieldUsageTransitive enables transitive method→field propagation.
	FieldUsageTransitive bool
	// ContinueOnError skips packages that fail to load or type-check.
	ContinueOnError bool
	// Weights configures the reusability components; the zero value selects
	// the defaults.
	Weights metrics.ReusabilityWeights
}

// TypeResult carries one type's reusability metric.
type TypeResult struct {
	// Name is the type's declared name.
	Name string
	// Reusability is the type-level reusability index result.
	Reusability metrics.MetricResult
}

// PackageResult carries one package's analyzed types.
type PackageResult struct {
	// Path is the package's import path.
	Path string
	// Types are the package's analyzed types, sorted by name.
	Types []TypeResult
}

// String summarizes the package result for debugging.
func (p PackageResult) String() string {
	return fmt.Sprintf("%s: %d types", p.Path, len(p.Types))
}

// Result is a deterministic analysis outcome: packages sorted by import
// path, types by name.
type Result struct {
	// ModulePath is the analyzed main module's path, when known.
	ModulePath string
	// Packages are the analyzed packages, sorted by import path.
	Packages []PackageResult
}

// Analyzer runs the full analysis pipeline. It implements no metric
// formula itself.
type Analyzer interface {
	Analyze(ctx context.Context, opts Options) (Result, error)
}
