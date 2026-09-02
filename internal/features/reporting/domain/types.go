// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"strings"

	"github.com/gostafa/reusability/reusability"
)

type (
	// DocScope classifies a metric guide entry's applicability.
	DocScope string

	// MetricDoc is one entry in the metrics guide.
	MetricDoc struct {
		// FormulaLaTeX is the metric formula in LaTeX.
		FormulaLaTeX string
		// NotApplicable explains when the metric does not apply.
		NotApplicable string
		// FullName is the human-readable metric title.
		FullName string
		// Scope classifies whether the metric is type- or package-level.
		Scope DocScope
		// Definition is the formula/version identifier for the metric.
		Definition string
		// FormulaMathML is the metric formula in MathML.
		FormulaMathML string
		// HowCalculated describes how the tool computes the value.
		HowCalculated string
		// Example is a short worked example of the metric.
		Example string
		// Label is the short column header for the metric.
		Label string
		// Interpretation explains how to read high versus low scores.
		Interpretation string
		// Name is the canonical metric identifier.
		Name string
		// Direction is whether higher or lower values are better.
		Direction string
		// Summary is a one-line description of the metric.
		Summary string
		// Bounded is true when the metric is normalized to [0, 1].
		Bounded bool
	}

	// Format names a report rendering format.
	Format string

	// TextOptions controls human-readable text report rendering.
	TextOptions struct {
		// Color wraps values in ANSI quality colors. Callers enable it only
		// when the destination understands escapes (a terminal).
		Color bool
		// Explain appends a notes section with the reasons behind n/a cells
		// and dropped metric components.
		Explain bool
	}

	scoreBias int

	metricQuality struct {
		bias    scoreBias
		bounded bool
	}

	tableCell struct {
		prefix string
		text   string
		style  string
	}

	treeNode struct {
		name     string
		pkg      *reusability.PackageReport
		children []*treeNode
	}

	treeSummary struct {
		typeAgg    map[string]*columnStats
		pkgsTotal  int
		typesTotal int
	}

	textTable struct {
		typeCols []string
		rows     [][]tableCell
	}

	columnStats struct {
		sum     float64
		count   int
		minimum float64
		maximum float64
	}

	noteEntry struct {
		reason string
		types  []string
	}

	paintArgs struct {
		text  string
		style string
		color bool
	}

	rowWriteArgs struct {
		builder *strings.Builder
		row     []tableCell
		widths  []int
		color   bool
	}

	notesBodyArgs struct {
		builder *strings.Builder
		path    string
		notes   []string
		color   bool
	}

	notesWriter struct {
		builder *strings.Builder
		color   bool
		wrote   bool
	}

	noteCollectState struct {
		index   map[string]int
		entries []noteEntry
		count   int
	}
)
