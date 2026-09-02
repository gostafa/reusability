// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package inbound

import (
	"context"

	"github.com/gostafa/reusability/internal/shared/metrics"
)

type (
	// Options configures one project analysis request.
	Options struct {
		Directory            string
		DependencyScope      string
		Patterns             []string
		BuildTags            []string
		Weights              metrics.ReusabilityWeights
		Workers              int
		IncludeTests         bool
		IncludeGenerated     bool
		FieldUsageTransitive bool
		ContinueOnError      bool
	}

	// TypeResult is one type's analysis outcome.
	TypeResult struct {
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
	Result struct {
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
