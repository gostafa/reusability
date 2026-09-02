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
		FormulaLaTeX   string
		NotApplicable  string
		FullName       string
		Scope          DocScope
		Definition     string
		FormulaMathML  string
		HowCalculated  string
		Example        string
		Label          string
		Interpretation string
		Name           string
		Direction      string
		Summary        string
		Bounded        bool
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

	emitArgs struct {
		node      *treeNode
		summaries map[*treeNode]*treeSummary
		prefix    string
		connector string
	}

	typeRowArgs struct {
		node      *treeNode
		summary   *treeSummary
		prefix    string
		connector string
		index     int
	}

	paintArgs struct {
		text  string
		style string
		color bool
	}

	footerArgs struct {
		report *reusability.Report
		opts   TextOptions
		sawNA  bool
	}

	branchEmitArgs struct {
		node        *treeNode
		summary     *treeSummary
		summaries   map[*treeNode]*treeSummary
		childPrefix string
		typeCount   int
		total       int
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

	cellPadArgs struct {
		row  *rowWriteArgs
		cell *tableCell
		idx  int
		last int
	}
)
