// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package inbound

import (
	"context"

	"github.com/gostafa/reusability/internal/shared/metrics"
)

type (
	// Options configures one project analysis request.
	Options = struct {
		// Directory is the module root used when resolving package patterns.
		Directory string
		// DependencyScope selects which imports count toward coupling
		// (project, module, or all).
		DependencyScope string
		// Patterns are go/packages patterns to load (for example ./...).
		Patterns []string
		// BuildTags are extra build tags passed to package loading.
		BuildTags []string
		// Weights are the reusability component weights for this run.
		Weights metrics.ReusabilityWeights
		// Workers is the max concurrent package workers (0 = auto).
		Workers int
		// IncludeTests includes test files and test packages when true.
		IncludeTests bool
		// IncludeGenerated includes generated source files when true.
		IncludeGenerated bool
		// FieldUsageTransitive counts transitive field usage for cohesion.
		FieldUsageTransitive bool
		// ContinueOnError skips packages that fail to load or type-check.
		ContinueOnError bool
	}

	// TypeResult is one type's analysis outcome.
	TypeResult = struct {
		// Name is the type's declared name.
		Name string
		// Reusability is the type-level reusability index result.
		Reusability metrics.MetricResult
	}

	// PackageResult is one package's analysis outcome.
	PackageResult struct {
		// Path is the package's import path.
		Path string
		// Types are the package's analyzed types, sorted by name.
		Types []TypeResult
	}

	// Result is the full project analysis outcome.
	Result = struct {
		// ModulePath is the analyzed main module's path, when known.
		ModulePath string
		// Packages are the analyzed packages, sorted by import path.
		Packages []PackageResult
	}

	// Analyzer runs the project analysis pipeline.
	Analyzer interface {
		Analyze(ctx context.Context, opts *Options) (Result, error)
	}
)
