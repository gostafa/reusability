// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"io"

	"github.com/gostafa/reusability/internal/features/reporting/domain"
	"github.com/gostafa/reusability/internal/features/reporting/ports/outbound"
	"github.com/gostafa/reusability/reusability"
)

type (
	// WriteRequest holds the inputs for Write.
	WriteRequest = struct {
		// Sink is the destination that receives the rendered report bytes.
		Sink outbound.Sink
		// Format selects the report rendering format (text, json, csv, or web).
		Format domain.Format
		// Report is the analysis report to render.
		Report reusability.Report
		// Options controls text-format rendering (color and explain notes).
		Options domain.TextOptions
	}

	renderInput = struct {
		writer io.Writer
		report *reusability.Report
		format domain.Format
		opts   domain.TextOptions
	}

	docsPayload = struct {
		// Tool identifies the producing tool.
		Tool jsonTool `json:"tool"`
		// Docs are the guide entries in render order.
		Docs []jsonMetricDoc `json:"docs"`
	}

	jsonMetricDoc = struct {
		FormulaLaTeX   string `json:"formula_latex,omitempty"`
		NotApplicable  string `json:"not_applicable,omitempty"`
		FullName       string `json:"full_name"`
		Scope          string `json:"scope"`
		Definition     string `json:"definition,omitempty"`
		FormulaMathML  string `json:"formula_mathml,omitempty"`
		How            string `json:"how"`
		Example        string `json:"example,omitempty"`
		Label          string `json:"label"`
		Interpretation string `json:"interpretation"`
		Name           string `json:"name"`
		Direction      string `json:"direction"`
		Summary        string `json:"summary"`
		Bounded        bool   `json:"bounded"`
	}

	jsonReport = struct {
		// SchemaVersion is the report schema version.
		SchemaVersion string `json:"schema_version"`
		// Tool identifies the producing tool.
		Tool jsonTool `json:"tool"`
		// Packages are the analyzed packages in report order.
		Packages []jsonPackage `json:"packages"`
	}

	jsonTool struct {
		// Name is the tool's canonical name.
		Name string `json:"name"`
		// Version is the tool's build version.
		Version string `json:"version"`
	}

	jsonPackage struct {
		// Path is the package's import path.
		Path string `json:"path"`
		// Types are the package's analyzed types in report order.
		Types []jsonType `json:"types"`
	}

	jsonType = struct {
		// Name is the type's declared name.
		Name string `json:"name"`
		// Reusability is the type-level reusability index result.
		Reusability jsonMetric `json:"reusability"`
	}

	jsonMetric = struct {
		Value      *float64 `json:"value,omitempty"`
		Reason     string   `json:"reason,omitempty"`
		Definition string   `json:"definition"`
		Applicable bool     `json:"applicable"`
	}

	webPayload = struct {
		// Module is the analyzed main module's path, when known.
		Module string `json:"module"`
		// Report is the same document the JSON format emits.
		Report jsonReport `json:"report"`
	}
)
