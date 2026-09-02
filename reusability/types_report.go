// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package reusability

import (
	"github.com/gostafa/reusability/internal/shared/metrics"
)

type (
	// MetricResult is one computed metric with applicability metadata.
	MetricResult = metrics.MetricResult

	// ToolInfo records the tool name and version that produced a report.
	ToolInfo struct {
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
	PackageReport struct {
		// Path is the package's import path.
		Path string
		// Types holds the package's analyzed named types, sorted by name.
		Types []TypeReport
	}

	// TypeReport is one named type's contribution to a PackageReport.
	TypeReport struct {
		// Name is the type's identifier within its package.
		Name string
		// Reusability is the type-level reusability index and its applicability
		// metadata (value, applicable, reason, definition).
		Reusability MetricResult
	}
)
