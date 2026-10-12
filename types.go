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

type (
	// MetricResult is one computed metric with applicability metadata.
	MetricResult = struct {
		Name       string  // metric identifier
		Scope      string  // type or package scope
		Reason     string  // applicability explanation
		Definition string  // formula version
		Value      float64 // computed score
		Applicable bool    // whether a score can be computed
	}

	// ToolInfo records the tool name and version that produced a report.
	ToolInfo = struct {
		// Name is the tool name embedded in reports; equals ToolName for this build.
		Name string
		// Version is the tool version string at analysis time.
		Version string
	}

	// Report is the deterministic analysis output.
	Report = struct {
		// SchemaVersion identifies the report schema; it equals the
		// SchemaVersion constant for reports this build produces.
		SchemaVersion string
		// Tool records the tool name and version that produced the report.
		Tool ToolInfo
		// Module is the analyzed main module's path, when known. Renderers
		// use it to show package paths relative to the module root.
		Module string
		// Packages holds one entry per analyzed package, sorted by import path.
		Packages []PackageReport
	}

	// PackageReport is one package's contribution to a Report.
	PackageReport = struct {
		// Path is the package's import path.
		Path string
		// Types holds the package's analyzed named types, sorted by name.
		Types []TypeReport
	}

	// TypeReport is one named type's contribution to a PackageReport.
	TypeReport = struct {
		// Name is the type's identifier within its package.
		Name string
		// Reusability is the type-level reusability index and its applicability
		// metadata (value, applicable, reason, definition).
		Reusability MetricResult
	}
)
