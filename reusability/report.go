package reusability

import "github.com/gostafa/reusability/internal/shared/metrics"

// SchemaVersion is the version of the report schema produced by Analyze.
// Version 6 keeps only package path, type name, and reusability per type.
const SchemaVersion = "6"

// ToolName is the canonical tool name embedded in reports.
const ToolName = "reusability"

// MetricResult aliases the metrics package's result type; see its
// documentation for the applicability contract.
type MetricResult = metrics.MetricResult

// ToolInfo identifies the tool that produced a report.
type ToolInfo struct {
	// Name is the tool name embedded in reports; equals ToolName for this build.
	Name string
	// Version is the tool version string at analysis time.
	Version string
}

// Report is the complete, deterministic result of one analysis run.
// Packages are sorted by import path; ordering never depends on map
// iteration.
type Report struct {
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

// PackageReport carries one package's analyzed types, sorted by name.
type PackageReport struct {
	// Path is the package's import path.
	Path string
	// Types holds the package's analyzed named types, sorted by name.
	Types []TypeReport
}

// TypeReport carries one named type's reusability result.
type TypeReport struct {
	// Name is the type's identifier within its package.
	Name string
	// Reusability is the type-level reusability index and its applicability
	// metadata (value, applicable, reason, definition).
	Reusability MetricResult
}
