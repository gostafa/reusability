package domain

import (
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/reusability"
)

// Format selects a report encoding.
type Format string

const (
	// FormatText renders a human-readable report.
	FormatText Format = "text"
	// FormatJSON renders the versioned JSON schema.
	FormatJSON Format = "json"
	// FormatCSV renders one row per entity and metric.
	FormatCSV Format = "csv"
	// FormatWeb renders a self-contained interactive HTML report.
	FormatWeb Format = "web"
)

// ParseFormat validates a format name.
func ParseFormat(name string) (Format, bool) {
	switch Format(name) {
	case FormatText, FormatJSON, FormatCSV, FormatWeb:
		return Format(name), true
	}

	return "", false
}

// FormatValue renders a metric value deterministically: the shortest
// decimal representation that round-trips, identical on every platform.
func FormatValue(value float64) string {
	return strconv.FormatFloat(value, 'g', -1, 64)
}

// TextOptions configures the text renderer.
type TextOptions struct {
	// Color wraps values in ANSI quality colors. Callers enable it only
	// when the destination understands escapes (a terminal).
	Color bool
	// Explain appends a notes section with the reasons behind n/a cells
	// and dropped metric components.
	Explain bool
}

const (
	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiDim    = "\x1b[2m"
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
)

// naCell marks a value that must not be read (metric not applicable).
const naCell = "–"

// metricQuality maps a metric's values to quality colors. Bounded metrics
// live in [0, 1] and use fixed thresholds; unbounded ones are judged
// relative to the other types in the same package. Metrics with no entry
// have no inherent good/bad direction.
type metricQuality struct {
	lowerBetter bool
	bounded     bool
}

var qualityByMetric = map[string]metricQuality{
	metrics.MetricReusability: {bounded: true},
}

// columnAbbrev holds the short table headings. Single-word metric names get
// a short, readable abbreviation; standard metric initialisms fall through
// to their uppercase name.
var columnAbbrev = map[string]string{
	metrics.MetricReusability: "Reuse",
}

func abbrev(name string) string {
	if short, ok := columnAbbrev[name]; ok {
		return short
	}

	return strings.ToUpper(name)
}

// tableCell is one table cell: an optional unstyled prefix (tree glyphs),
// the visible text, and its ANSI style.
type tableCell struct {
	prefix string
	text   string
	style  string
}

func (c tableCell) width() int {
	return utf8.RuneCountInString(c.prefix) + utf8.RuneCountInString(c.text)
}

// treeNode is one node of the package tree: a path segment (or several when
// single-child chains are compressed) and, for package nodes, its report.
type treeNode struct {
	name     string
	pkg      *reusability.PackageReport
	children []*treeNode
}

// treeSummary holds the computed statistics for one node's complete subtree.
// Keeping it separate leaves treeNode responsible only for tree topology.
type treeSummary struct {
	pkgsTotal   int
	typesTotal  int // types in the subtree, applicable or not
	funcsTotal  int
	varsTotal   int
	constsTotal int
	typeAgg     map[string]*columnStats // subtree type metrics
}

func (n *treeNode) child(name string) *treeNode {
	for _, c := range n.children {
		if c.name == name {
			return c
		}
	}

	c := &treeNode{name: name}
	n.children = append(n.children, c)

	return c
}

// aggregateTree computes the mean inputs and value ranges for every metric in
// n's subtree and indexes each node's result for the rendering pass.
func aggregateTree(n *treeNode, summaries map[*treeNode]*treeSummary) *treeSummary {
	summary := &treeSummary{typeAgg: make(map[string]*columnStats)}
	summaries[n] = summary

	if n.pkg != nil {
		summary.pkgsTotal = 1
		summary.typesTotal = len(n.pkg.Types)
		summary.funcsTotal = n.pkg.ExportedFuncs + n.pkg.UnexportedFuncs
		summary.varsTotal = n.pkg.Vars
		summary.constsTotal = n.pkg.Consts
		collectPackageStats(n.pkg, summary.typeAgg)
	}

	for _, c := range n.children {
		childSummary := aggregateTree(c, summaries)
		summary.pkgsTotal += childSummary.pkgsTotal
		summary.typesTotal += childSummary.typesTotal
		summary.funcsTotal += childSummary.funcsTotal
		summary.varsTotal += childSummary.varsTotal
		summary.constsTotal += childSummary.constsTotal
		mergeStats(summary.typeAgg, childSummary.typeAgg)
	}

	return summary
}

// collectPackageStats feeds one package's applicable type-metric values into
// the aggregation map.
func collectPackageStats(pkg *reusability.PackageReport, typeAgg map[string]*columnStats) {
	for i := range pkg.Types {
		for _, r := range pkg.Types[i].Metrics {
			if r.Applicable {
				addStat(typeAgg, r.Name, r.Value)
			}
		}
	}
}

func addStat(m map[string]*columnStats, name string, value float64) {
	st := m[name]
	if st == nil {
		st = &columnStats{min: value, max: value}
		m[name] = st
	}

	st.sum += value
	st.count++
	st.min = math.Min(st.min, value)
	st.max = math.Max(st.max, value)
}

func mergeStats(dst, src map[string]*columnStats) {
	for name, s := range src {
		d := dst[name]
		if d == nil {
			c := *s
			dst[name] = &c

			continue
		}

		d.sum += s.sum
		d.count += s.count
		d.min = math.Min(d.min, s.min)
		d.max = math.Max(d.max, s.max)
	}
}

// compress merges directory chains with a single child and no package of
// their own, so "internal" → "features" renders as "internal/features".
func (n *treeNode) compress() {
	for n.pkg == nil && len(n.children) == 1 {
		c := n.children[0]
		n.name = n.name + "/" + c.name
		n.pkg = c.pkg
		n.children = c.children
	}

	for _, c := range n.children {
		c.compress()
	}
}

// buildTree arranges the report's packages by their module-relative paths.
// The returned root carries the module-root package (if analyzed) and one
// child per top-level path segment.
func buildTree(report reusability.Report) *treeNode {
	root := &treeNode{}

	for i := range report.Packages {
		pkg := &report.Packages[i]

		rel := relPath(pkg.Path, report.Module)
		if rel == "." {
			root.pkg = pkg

			continue
		}

		node := root
		for seg := range strings.SplitSeq(rel, "/") {
			node = node.child(seg)
		}

		node.pkg = pkg
	}

	for _, c := range root.children {
		c.compress()
	}

	return root
}

func relPath(path, module string) string {
	if module == "" {
		return path
	}

	if path == module {
		return "."
	}

	if strings.HasPrefix(path, module+"/") {
		return path[len(module)+1:]
	}

	return path
}

// Text renders the whole report as one tree table: the module-root package row
// summarizes the complete module, leaf package rows carry their exact metrics,
// and other parent/group rows carry the means of all package metrics below
// them. Every group mean uses applicable values only. Types follow as branches
// with their metric columns. With Explain, the reasons behind n/a cells follow
// as a notes section.
func Text(report reusability.Report, opts TextOptions) string {
	var b strings.Builder
	b.WriteString(report.Tool.Name)
	b.WriteString(" ")
	b.WriteString(report.Tool.Version)
	b.WriteString(" — schema ")
	b.WriteString(report.SchemaVersion)
	b.WriteString("\n")

	if report.Module != "" {
		b.WriteString("module ")
		b.WriteString(paint(report.Module, ansiBold, opts.Color))
		b.WriteString("\n")
	}

	if len(report.Packages) == 0 {
		return b.String()
	}

	b.WriteString("\n")

	table := &textTable{typeCols: reportColumns(report)}
	summaries := make(map[*treeNode]*treeSummary)

	header := make([]tableCell, 0, 1+len(table.typeCols))

	header = append(header, tableCell{text: "PATH / TYPE", style: ansiDim})
	for _, name := range structuralPackageColumns() {
		header = append(header, tableCell{text: name, style: ansiDim})
	}
	for _, name := range table.typeCols {
		header = append(header, tableCell{text: abbrev(name), style: ansiDim})
	}

	table.rows = append(table.rows, header)

	root := buildTree(report)
	root.name = "."
	aggregateTree(root, summaries)

	if root.pkg != nil {
		table.emitModuleSummary(root, summaries[root])
	}

	for i, section := range root.children {
		if root.pkg != nil || i > 0 {
			table.rows = append(table.rows, nil)
		}

		table.emitNode(section, summaries, "", "")
	}

	// Global column widths keep every branch aligned as one table.
	sawNA := false

	widths := make([]int, 1+len(structuralPackageColumns())+len(table.typeCols))
	for _, row := range table.rows {
		for c, cell := range row {
			widths[c] = max(widths[c], cell.width())
			if cell.text == naCell {
				sawNA = true
			}
		}
	}

	for _, row := range table.rows {
		if len(row) == 0 {
			b.WriteString("\n")

			continue
		}

		last := len(row) - 1
		for last > 0 && row[last].text == "" && row[last].prefix == "" {
			last--
		}

		for c, cell := range row[:last+1] {
			b.WriteString(cell.prefix)
			b.WriteString(paint(cell.text, cell.style, opts.Color))

			if c < last {
				b.WriteString(strings.Repeat(" ", widths[c]-cell.width()+2))
			}
		}

		b.WriteString("\n")
	}

	if sawNA {
		b.WriteString("\n")
		b.WriteString(paint(naCell+" = not applicable", ansiDim, opts.Color))
		b.WriteString("\n")
	}

	if opts.Explain {
		writeNotes(&b, report, opts.Color)
	}

	return b.String()
}

// textTable accumulates the tree table's rows before column widths are
// known. A nil row renders as a blank separator line.
type textTable struct {
	typeCols []string
	rows     [][]tableCell
}

// emitModuleSummary appends the module-wide root row and any types declared by
// the root package. Top-level package sections are emitted separately.
func (t *textTable) emitModuleSummary(node *treeNode, summary *treeSummary) {
	t.nodeRow(node, summary, "")

	typeCount := nodeTypeCount(node, t.typeCols)
	for i := range typeCount {
		t.typeRow(node, summary, i, "", branchGlyph(i, typeCount))
	}
}

// naTableCell renders a value that must not be read.
func naTableCell() tableCell {
	return tableCell{text: naCell, style: ansiDim}
}

// emitNode appends one package or directory node and its subtree: the node
// row, its type branches, then child nodes, drawing the tree glyphs.
func (t *textTable) emitNode(
	node *treeNode,
	summaries map[*treeNode]*treeSummary,
	prefix, connector string,
) {
	t.nodeRow(node, summaries[node], prefix+connector)

	childPrefix := childIndent(prefix, connector)
	typeCount := nodeTypeCount(node, t.typeCols)
	total := typeCount + len(node.children)

	for i := range typeCount {
		t.typeRow(node, summaries[node], i, childPrefix, branchGlyph(i, total))
	}

	for i, child := range node.children {
		if typeCount+i > 0 {
			t.rows = append(t.rows, []tableCell{{prefix: childPrefix + "│"}})
		}

		t.emitNode(child, summaries, childPrefix, branchGlyph(typeCount+i, total))
	}
}

// nodeTypeCount is the number of type rows a node contributes: its package's
// types when any type column is displayed, otherwise none.
func nodeTypeCount(node *treeNode, typeCols []string) int {
	if node.pkg != nil && len(typeCols) > 0 {
		return len(node.pkg.Types)
	}

	return 0
}

// branchGlyph is the tree connector for the child at index of total: the
// corner glyph for the last child, a tee for the rest.
func branchGlyph(index, total int) string {
	if index == total-1 {
		return "└── "
	}

	return "├── "
}

// childIndent extends a node's prefix for its children, continuing the
// vertical guide under a tee and leaving blank space under a corner.
func childIndent(prefix, connector string) string {
	switch connector {
	case "├── ":
		return prefix + "│   "
	case "└── ":
		return prefix + "    "
	}

	return prefix
}

// nodeRow appends the spanning row of one package or directory node: its
// structural counts followed by the means over all types in its subtree.
func (t *textTable) nodeRow(node *treeNode, summary *treeSummary, label string) {
	row := make([]tableCell, 0, 1+len(structuralPackageColumns())+len(t.typeCols))
	row = append(row, tableCell{prefix: label, text: node.name, style: ansiBold})
	row = append(row, structuralPackageCells(node, summary)...)
	row = append(row, nodeTypeAggCells(summary, t.typeCols)...)

	t.rows = append(t.rows, row)
}

func structuralPackageColumns() []string {
	return []string{"Funcs", "Vars", "Consts", "Types"}
}

func structuralPackageCells(node *treeNode, summary *treeSummary) []tableCell {
	funcs := summary.funcsTotal
	vars := summary.varsTotal
	consts := summary.constsTotal
	types := summary.typesTotal
	if node.pkg != nil && len(node.children) == 0 {
		funcs = node.pkg.ExportedFuncs + node.pkg.UnexportedFuncs
		vars = node.pkg.Vars
		consts = node.pkg.Consts
		types = len(node.pkg.Types)
	}

	return []tableCell{
		{text: strconv.Itoa(funcs), style: ansiBold},
		{text: strconv.Itoa(vars), style: ansiBold},
		{text: strconv.Itoa(consts), style: ansiBold},
		{text: strconv.Itoa(types), style: ansiBold},
	}
}

// nodeTypeAggCells renders a node's type columns as the means over all types
// in its subtree; empty when the subtree holds no types.
func nodeTypeAggCells(summary *treeSummary, typeCols []string) []tableCell {
	if summary.typesTotal == 0 {
		return nil
	}

	cells := make([]tableCell, 0, len(typeCols))
	for _, name := range typeCols {
		st := summary.typeAgg[name]
		cells = append(
			cells,
			meanCell(st, func(v float64) string { return valueColor(name, v, st) }),
		)
	}

	return cells
}

// typeRow appends one type's branch row with its metric values, colored
// against the subtree's column ranges.
func (t *textTable) typeRow(
	node *treeNode,
	summary *treeSummary,
	index int,
	prefix, connector string,
) {
	typ := &node.pkg.Types[index]
	byName := metricsByName(typ.Metrics)

	row := make([]tableCell, 0, 1+len(structuralPackageColumns())+len(t.typeCols))
	row = append(row, tableCell{prefix: prefix + connector, text: typ.Name})

	for range structuralPackageColumns() {
		row = append(row, tableCell{})
	}

	row = append(row, typeMetricCells(byName, t.typeCols, summary.typeAgg)...)

	t.rows = append(t.rows, row)
}

// typeMetricCells renders one type's metric columns, coloring each applicable
// value against the subtree's column range and marking the rest n/a.
func typeMetricCells(
	byName map[string]metrics.MetricResult,
	typeCols []string,
	stats map[string]*columnStats,
) []tableCell {
	cells := make([]tableCell, 0, len(typeCols))
	for _, name := range typeCols {
		r, ok := byName[name]
		if !ok || !r.Applicable {
			cells = append(cells, naTableCell())

			continue
		}

		cells = append(cells, tableCell{
			text:  formatCell(r.Value),
			style: valueColor(name, r.Value, stats[name]),
		})
	}

	return cells
}

// metricsByName indexes metric results by their metric name.
func metricsByName(results []metrics.MetricResult) map[string]metrics.MetricResult {
	byName := make(map[string]metrics.MetricResult, len(results))
	for _, r := range results {
		byName[r.Name] = r
	}

	return byName
}

// meanCell renders one aggregated column: the mean of the applicable
// values, or the n/a marker when none exist.
func meanCell(st *columnStats, color func(float64) string) tableCell {
	if st == nil || st.count == 0 {
		return naTableCell()
	}

	value := st.sum / float64(st.count)

	return tableCell{text: formatCell(value), style: ansiBold + color(value)}
}

// boundedColorFor adapts boundedColor to the meanCell color callback.
func boundedColorFor(name string) func(float64) string {
	return func(value float64) string { return boundedColor(name, value) }
}

// columnStats aggregates one metric column: the mean input and the value
// range used for relative coloring of unbounded metrics.
type columnStats struct {
	sum   float64
	count int
	min   float64
	max   float64
}

// reportColumns lists the type-level metrics present anywhere in the
// report, in the fixed metric order.
func reportColumns(report reusability.Report) []string {
	present := make(map[string]bool)

	for i := range report.Packages {
		for j := range report.Packages[i].Types {
			for _, r := range report.Packages[i].Types[j].Metrics {
				present[r.Name] = true
			}
		}
	}

	var cols []string

	for _, name := range metrics.TypeMetricOrder() {
		if present[name] {
			cols = append(cols, name)
		}
	}

	return cols
}

// writeNotes appends the notes section: per package, the reasons dropped
// from table cells.
func writeNotes(b *strings.Builder, report reusability.Report, color bool) {
	wrote := false

	for i := range report.Packages {
		pkg := &report.Packages[i]

		notes := packageNotes(pkg)
		if len(notes) == 0 {
			continue
		}

		if !wrote {
			b.WriteString("\n")
			b.WriteString(paint("notes", ansiDim, color))
			b.WriteString("\n")

			wrote = true
		}

		b.WriteString("  ")
		b.WriteString(paint(pkg.Path, ansiDim, color))
		b.WriteString("\n")

		for _, note := range notes {
			b.WriteString("    ")
			b.WriteString(paint(note, ansiDim, color))
			b.WriteString("\n")
		}
	}
}

// packageNotes collects the reasons dropped from table cells: n/a
// explanations and component notes. Identical reasons are aggregated into
// one line per metric, listing the affected types — otherwise packages
// full of method-less types bury the table under repeated boilerplate.
func packageNotes(pkg *reusability.PackageReport) []string {
	var notes []string

	for _, name := range metrics.TypeMetricOrder() {
		type entry struct {
			reason string
			types  []string
		}

		var entries []entry

		index := make(map[string]int)

		for i := range pkg.Types {
			for _, r := range pkg.Types[i].Metrics {
				if r.Name != name || r.Reason == "" {
					continue
				}

				j, ok := index[r.Reason]
				if !ok {
					j = len(entries)
					index[r.Reason] = j
					entries = append(entries, entry{reason: r.Reason})
				}

				entries[j].types = append(entries[j].types, pkg.Types[i].Name)
			}
		}

		for _, e := range entries {
			who := strings.Join(e.types, ", ")
			if len(e.types) == len(pkg.Types) && len(pkg.Types) > 1 {
				who = "all types"
			}

			notes = append(notes, name+": "+e.reason+" ("+who+")")
		}
	}

	return notes
}

// formatCell renders a value for a table cell with two decimals, keeping
// columns uniform. Machine formats keep FormatValue.
func formatCell(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

// valueColor picks the quality color for one value of the named metric.
// Unbounded metrics are normalized against their package column's range;
// a column with no spread (or a lone type) stays plain.
func valueColor(name string, value float64, st *columnStats) string {
	q, ok := qualityByMetric[name]
	if !ok {
		return ""
	}

	if !q.bounded {
		if st == nil || st.max == st.min {
			return ""
		}

		value = (value - st.min) / (st.max - st.min)
	}

	return thresholdColor(q.lowerBetter, value)
}

// boundedColor colors package-row values, where no column context exists;
// only metrics with fixed 0–1 thresholds get a color.
func boundedColor(name string, value float64) string {
	q, ok := qualityByMetric[name]
	if !ok || !q.bounded {
		return ""
	}

	return thresholdColor(q.lowerBetter, value)
}

func thresholdColor(lowerBetter bool, score float64) string {
	if lowerBetter {
		score = 1 - score
	}

	switch {
	case score >= 0.66:
		return ansiGreen
	case score >= 0.33:
		return ansiYellow
	default:
		return ansiRed
	}
}

// paint wraps text in an ANSI style when coloring is enabled.
func paint(text, style string, enabled bool) string {
	if !enabled || style == "" {
		return text
	}

	return style + text + ansiReset
}

// CSVHeader is the fixed CSV column set.
func CSVHeader() []string {
	return []string{
		"package",
		"type",
		"metric",
		"scope",
		"value",
		"applicable",
		"reason",
		"definition",
	}
}

// CSVRecords flattens the report into one record per entity and metric, in
// report order.
func CSVRecords(report reusability.Report) [][]string {
	var records [][]string

	appendRecords := func(pkgPath, typeName string, results []metrics.MetricResult) {
		for _, r := range results {
			value := ""
			if r.Applicable {
				value = FormatValue(r.Value)
			}

			records = append(records, []string{
				pkgPath, typeName, r.Name, string(r.Scope), value,
				strconv.FormatBool(r.Applicable), r.Reason, r.Definition,
			})
		}
	}

	for i := range report.Packages {
		pkg := &report.Packages[i]

		for j := range pkg.Types {
			appendRecords(pkg.Path, pkg.Types[j].Name, pkg.Types[j].Metrics)
		}
	}

	return records
}
