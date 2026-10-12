// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

// Package reusability provides reusable analysis utilities.
package reusability

type (
	// DependencyScope selects which imports count toward coupling.
	DependencyScope = string

	// FieldUsageMode selects direct or transitive field-usage counting.
	FieldUsageMode = string

	// MetricName identifies a reported metric.
	MetricName = string

	// Weights is the reusability component weight set.
	Weights = struct {
		Cohesion      float64 // weight of the cohesion component
		Coupling      float64 // weight of the coupling component
		Testability   float64 // weight of the testability component
		Documentation float64 // weight of the documentation component
	}

	// Config configures one Analyze invocation.
	Config = struct {
		// Directory is the module root used for package loading.
		Directory string
		// DependencyScope selects which imports count toward coupling
		// (project, module, or all).
		DependencyScope DependencyScope
		// FieldUsageMode selects direct or transitive field-usage counting.
		FieldUsageMode FieldUsageMode
		// Patterns are go/packages patterns to analyze (default ./...).
		Patterns []string
		// BuildTags are extra build tags passed to package loading.
		BuildTags []string
		// ReusabilityWeights are the four component weights for the index.
		ReusabilityWeights Weights
		// Workers is the max concurrent type workers (0 picks a default).
		Workers int
		// IncludeTests includes *_test.go packages when true.
		IncludeTests bool
		// IncludeGenerated includes generated files when true.
		IncludeGenerated bool
		// ContinueOnError keeps analyzing after package load errors.
		ContinueOnError bool
	}
)
