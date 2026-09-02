package reusability

import "github.com/gostafa/reusability/internal/shared/metrics"

// SchemaVersion is the version of the report schema produced by Analyze.
// Version 3 added package-level vars and consts counts.
// Version 4 added declaration details plus per-function complexity and lines.
const SchemaVersion = "4"

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

// PackageReport carries one package's structural facts and its analyzed
// types, sorted by name. Package-level metrics are not reported.
type PackageReport struct {
	// Path is the package's import path.
	Path string
	// Afferent counts analyzed packages importing this package (Ca).
	Afferent int
	// Efferent counts this package's in-scope imports (Ce), honoring the
	// configured dependency scope.
	Efferent int
	// ExportedFuncs counts the package's declared functions and methods with
	// an exported name.
	ExportedFuncs int
	// UnexportedFuncs counts the package's declared functions and methods with
	// an unexported name.
	UnexportedFuncs int
	// Vars counts top-level variable names declared in the package.
	Vars int
	// Consts counts top-level constant names declared in the package.
	Consts int
	// Variables holds top-level variable declarations in source order.
	Variables []DeclarationReport
	// Constants holds top-level constant declarations in source order.
	Constants []DeclarationReport
	// Functions holds top-level functions in source order, excluding methods.
	Functions []FunctionReport
	// Types holds the package's analyzed named types, sorted by name.
	Types []TypeReport
}

// TypeReport carries one named type's structural facts and its metrics in
// the fixed metric order, restricted to the selected display set.
type TypeReport struct {
	// Name is the type's identifier within its package.
	Name string
	// Exported reports whether the type name is exported.
	Exported bool
	// Kind classifies the type's underlying type.
	Kind string
	// Position locates the type declaration in source.
	Position Position
	// Fields is the struct field count (embedded fields count one).
	Fields int
	// FieldDetails holds the struct fields in declaration order.
	FieldDetails []FieldReport
	// Methods is the declared method count.
	Methods int
	// MethodDetails holds the declared methods in report order.
	MethodDetails []FunctionReport
	// Metrics holds the type-level metric results in the fixed metric order,
	// restricted to the selected display set.
	Metrics []MetricResult
}

// Position locates a declaration in source.
type Position struct {
	// File is the source file path, relative when possible.
	File string
	// Line is the 1-based source line.
	Line int
	// Column is the 1-based source column.
	Column int
}

// DeclarationReport carries one top-level variable or constant declaration.
type DeclarationReport struct {
	// Name is the declared identifier.
	Name string
	// Exported reports whether the declaration name is exported.
	Exported bool
	// Position locates the declaration in source.
	Position Position
}

// FieldReport describes one struct field slot.
type FieldReport struct {
	// Name is the field name (the type name for embedded fields).
	Name string
	// Exported reports whether the field name is exported.
	Exported bool
	// Embedded marks an embedded field.
	Embedded bool
}

// FunctionReport carries one function or method's structural details.
type FunctionReport struct {
	// Name is the declared function or method name.
	Name string
	// Exported reports whether the function name is exported.
	Exported bool
	// Receiver is the declaring type name for methods; empty for free funcs.
	Receiver string
	// Position locates the declaration in source.
	Position Position
	// Lines is the inclusive source line count from func keyword to end.
	Lines int
	// Cyclomatic is the cyclomatic complexity score.
	Cyclomatic int
	// Branches are the syntax counts feeding cyclomatic complexity.
	Branches BranchStats
}

// BranchStats counts the syntax constructs that increment cyclomatic
// complexity.
type BranchStats struct {
	Ifs         int
	Fors        int
	Ranges      int
	Cases       int
	SelectComms int
	LogicalOps  int
}
