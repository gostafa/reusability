// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/reusability"
)

// CSVHeader is the fixed CSV column set.
func CSVHeader() []string {
	return []string{
		"package",
		"type",
		metrics.MetricReusability,
		"applicable",
		"reason",
		"definition",
	}
}

// CSVRecords flattens the report into one row per type, in report order.
func CSVRecords(report *reusability.Report) [][]string {
	total := indexZero

	for i := range report.Packages {
		total += len(report.Packages[i].Types)
	}

	records := make([][]string, indexZero, total)

	for i := range report.Packages {
		records = append(records, packageCSVRecords(&report.Packages[i])...)
	}

	return records
}

func packageCSVRecords(pkg *reusability.PackageReport) [][]string {
	records := make([][]string, indexZero, len(pkg.Types))

	for idx := range pkg.Types {
		records = append(records, typeCSVRecord(pkg.Path, &pkg.Types[idx]))
	}

	return records
}

func typeCSVRecord(pkgPath string, typ *reusability.TypeReport) []string {
	reuse := typ.Reusability
	value := emptyString

	if reuse.Applicable {
		value = FormatValue(reuse.Value)
	}

	return []string{
		pkgPath, typ.Name, value,
		strconv.FormatBool(reuse.Applicable), reuse.Reason, reuse.Definition,
	}
}

// FormatValue renders a metric value deterministically: the shortest
// decimal representation that round-trips, identical on every platform.
func FormatValue(value float64) string {
	return strconv.FormatFloat(value, 'g', -1, floatBits)
}

// MetricDocs returns the guide entries for the reported metric. Internal
// inputs (AMC, LCOM, TCC, CBO) are described in the reusability formula;
// they are not documented as selectable metrics.
func MetricDocs() []MetricDoc {
	return []MetricDoc{{
		Name:           metrics.MetricReusability,
		Label:          abbrev(metrics.MetricReusability),
		FullName:       "Experimental Reusability Index",
		Scope:          DocScopeType,
		Definition:     metrics.DefinitionReusability,
		FormulaMathML:  formulaReusability(),
		FormulaLaTeX:   formulaLaTeXReusability,
		Summary:        summaryReusability,
		HowCalculated:  howCalculatedReusability,
		Interpretation: interpretationReusability,
		NotApplicable:  notApplicableReusability,
		Direction:      DirectionHigher,
		Bounded:        true,
		Example:        exampleReusability,
	}}
}

// ParseFormat validates a format name.
func ParseFormat(name string) (Format, bool) {
	switch Format(name) {
	case FormatText, FormatJSON, FormatCSV, FormatWeb:
		return Format(name), true
	default:
		return emptyString, false
	}
}

// Text renders the whole report as one tree table: the module-root package row
// summarizes the complete module, leaf package rows carry their exact metrics,
// and other parent/group rows carry the means of all package metrics below
// them. Every group mean uses applicable values only. Types follow as branches
// with their metric columns. With Explain, the reasons behind n/a cells follow
// as a notes section.
func Text(report *reusability.Report, opts TextOptions) (string, error) {
	var builder strings.Builder

	err := writeTextReport(&builder, report, opts)
	if err != nil {
		return emptyString, fmt.Errorf("write text report: %w", err)
	}

	return builder.String(), nil
}

func writeTextReport(builder *strings.Builder, report *reusability.Report, opts TextOptions) error {
	err := writeTextHeader(builder, report, opts.Color)
	if err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	if len(report.Packages) == indexZero {
		return nil
	}

	err = writeTextTableBody(builder, report, opts)
	if err != nil {
		return fmt.Errorf("write table body: %w", err)
	}

	return nil
}

func writeTextTableBody(
	builder *strings.Builder,
	report *reusability.Report,
	opts TextOptions,
) error {
	// Build and emit the text table body.
	table := buildTextTable(report)

	sawNA, err := writeTextRows(builder, table, opts.Color)
	if err != nil {
		return fmt.Errorf("write rows: %w", err)
	}

	err = writeTextFooter(builder, report, opts, sawNA)
	if err != nil {
		return fmt.Errorf("write footer: %w", err)
	}

	return nil
}

func writeTextHeader(builder *strings.Builder, report *reusability.Report, color bool) error {
	err := writeBuilderStrings(
		builder,
		report.Tool.Name,
		spaceString,
		report.Tool.Version,
		" — schema ",
		report.SchemaVersion,
		newline,
	)
	if err != nil {
		return fmt.Errorf("write tool line: %w", err)
	}

	err = writeModuleLine(builder, report.Module, color)
	if err != nil {
		return fmt.Errorf("write module line: %w", err)
	}

	return nil
}

func writeModuleLine(builder *strings.Builder, module string, color bool) error {
	if module == emptyString {
		return nil
	}

	err := writeBuilderStrings(
		builder,
		"module ",
		paint(&paintArgs{text: module, style: ansiBold, color: color}),
		newline,
	)
	if err != nil {
		return fmt.Errorf("write module: %w", err)
	}

	return nil
}

func buildTextTable(report *reusability.Report) *textTable {
	table := &textTable{typeCols: reportColumns(report)}

	table.rows = append(table.rows, textHeaderRow(table.typeCols))

	root := buildTree(report)

	root.name = pathDot

	summaries := make(map[*treeNode]*treeSummary)
	aggregateTree(root, summaries)
	emitTreeRows(table, root, summaries)

	return table
}

func textHeaderRow(typeCols []string) []tableCell {
	header := make([]tableCell, indexZero, len(typeCols)+1)

	header = append(header, tableCell{text: pathTypeHeader, style: ansiDim})

	for i := range typeCols {
		header = append(header, tableCell{text: abbrev(typeCols[i]), style: ansiDim})
	}

	return header
}

func emitTreeRows(table *textTable, root *treeNode, summaries map[*treeNode]*treeSummary) {
	if root.pkg != nil {
		emitModuleSummary(table, root, summaries[root])
	}

	for i := range root.children {
		if root.pkg != nil || i > indexZero {
			table.rows = append(table.rows, nil)
		}

		emitNode(table, root.children[i], summaries, emptyString, emptyString)
	}
}

func writeTextFooter(
	builder *strings.Builder,
	report *reusability.Report,
	opts TextOptions,
	sawNA bool,
) error {
	err := writeNALegend(builder, opts.Color, sawNA)
	if err != nil {
		return fmt.Errorf("write na legend: %w", err)
	}

	err = writeExplainNotes(builder, report, opts)
	if err != nil {
		return fmt.Errorf("write explain notes: %w", err)
	}

	return nil
}

func writeNALegend(builder *strings.Builder, color, sawNA bool) error {
	if !sawNA {
		return nil
	}

	err := writeBuilderStrings(
		builder,
		newline,
		paint(&paintArgs{
			text: naCell + " = not applicable", style: ansiDim, color: color,
		}),
		newline,
	)
	if err != nil {
		return fmt.Errorf("write na strings: %w", err)
	}

	return nil
}

func writeExplainNotes(
	builder *strings.Builder,
	report *reusability.Report,
	opts TextOptions,
) error {
	if !opts.Explain {
		return nil
	}

	err := writeNotes(builder, report, opts.Color)
	if err != nil {
		return fmt.Errorf("write notes: %w", err)
	}

	return nil
}

func writeTextRows(builder *strings.Builder, table *textTable, color bool) (bool, error) {
	widths, sawNA := measureRows(table)

	for i := range table.rows {
		err := writeOneRow(&rowWriteArgs{
			builder: builder, row: table.rows[i], widths: widths, color: color,
		})
		if err != nil {
			return false, fmt.Errorf(errWrapWriteRow, err)
		}
	}

	return sawNA, nil
}

func measureRows(table *textTable) ([]int, bool) {
	widths := zeroIntSlice(len(table.typeCols) + 1)
	sawNA := false

	for i := range table.rows {
		sawNA = measureRow(table.rows[i], widths) || sawNA
	}

	return widths, sawNA
}

func zeroIntSlice(size int) []int {
	out := make([]int, indexZero, size)

	for range size {
		out = append(out, indexZero)
	}

	return out
}

func zeroNoteEntries(size int) []noteEntry {
	out := make([]noteEntry, indexZero, size)

	for range size {
		out = append(out, noteEntry{})
	}

	return out
}

func measureRow(row []tableCell, widths []int) bool {
	sawNA := false

	for idx := range row {
		widths[idx] = max(widths[idx], row[idx].width())

		if row[idx].text == naCell {
			sawNA = true
		}
	}

	return sawNA
}

func writeOneRow(args *rowWriteArgs) error {
	writer := writeBlankRowChecked

	if len(args.row) != indexZero {
		writer = writeFilledRowChecked
	}

	err := writer(args)
	if err != nil {
		return fmt.Errorf(errWrapWriteRow, err)
	}

	return nil
}

func writeBlankRowChecked(args *rowWriteArgs) error {
	err := writeBlankRow(args)
	if err != nil {
		return fmt.Errorf("blank row: %w", err)
	}

	return nil
}

func writeFilledRowChecked(args *rowWriteArgs) error {
	err := writeFilledRow(args)
	if err != nil {
		return fmt.Errorf("filled row: %w", err)
	}

	return nil
}

func writeBlankRow(args *rowWriteArgs) error {
	err := writeBuilderString(args.builder, newline)
	if err != nil {
		return fmt.Errorf("write blank row: %w", err)
	}

	return nil
}

func writeFilledRow(args *rowWriteArgs) error {
	last := lastFilledCell(args.row)

	err := writeRowCells(args, last)
	if err != nil {
		return fmt.Errorf("write cells: %w", err)
	}

	err = writeBuilderString(args.builder, newline)
	if err != nil {
		return fmt.Errorf("write row end: %w", err)
	}

	return nil
}

func writeRowCells(args *rowWriteArgs, last int) error {
	for idx := range args.row[:last+countOne] {
		err := writeRowCell(args, idx, last)
		if err != nil {
			return fmt.Errorf("write row cell: %w", err)
		}
	}

	return nil
}

func writeRowCell(args *rowWriteArgs, idx, last int) error {
	cell := &args.row[idx]

	err := writeBuilderStrings(
		args.builder,
		cell.prefix,
		paint(&paintArgs{text: cell.text, style: cell.style, color: args.color}),
	)
	if err != nil {
		return fmt.Errorf("write cell: %w", err)
	}

	err = writeCellPad(args, cell, idx, last)
	if err != nil {
		return fmt.Errorf("write cell pad: %w", err)
	}

	return nil
}

func writeCellPad(args *rowWriteArgs, cell *tableCell, idx, last int) error {
	if idx >= last {
		return nil
	}

	pad := strings.Repeat(spaceString, args.widths[idx]-cell.width()+countTwo)

	err := writeBuilderString(args.builder, pad)
	if err != nil {
		return fmt.Errorf("write pad: %w", err)
	}

	return nil
}

func lastFilledCell(row []tableCell) int {
	last := len(row) - countOne

	for last > indexZero && row[last].text == emptyString && row[last].prefix == emptyString {
		last--
	}

	return last
}

func abbrev(name string) string {
	if short, ok := columnAbbrev(name); ok {
		return short
	}

	return strings.ToUpper(name)
}

func columnAbbrev(name string) (string, bool) {
	if name == metrics.MetricReusability {
		return "Reuse", true
	}

	return emptyString, false
}

func formulaReusability() string {
	return mathMLReusabilityRI + newline +
		mathMLReusabilityC + newline +
		mathMLReusabilityK + newline +
		mathMLReusabilityT + newline +
		mathMLReusabilityD
}

func qualityForMetric(name string) (metricQuality, bool) {
	if name == metrics.MetricReusability {
		return metricQuality{bounded: true, bias: biasHigherBetter}, true
	}

	return metricQuality{}, false
}

func writeBuilderString(builder *strings.Builder, text string) error {
	written, err := builder.WriteString(text)
	if err != nil {
		return fmt.Errorf("builder write: %w", err)
	}

	if written != len(text) {
		return fmt.Errorf("builder write: wrote %d of %d: %w", written, len(text), errShortWrite)
	}

	return nil
}

func writeBuilderStrings(builder *strings.Builder, parts ...string) error {
	for i := range parts {
		err := writeBuilderString(builder, parts[i])
		if err != nil {
			return fmt.Errorf("builder strings: %w", err)
		}
	}

	return nil
}

func addStat(stats map[string]*columnStats, name string, value float64) {
	stat := stats[name]

	if stat == nil {
		stat = &columnStats{minimum: value, maximum: value}
		stats[name] = stat
	}

	stat.sum += value
	stat.count++

	stat.minimum = math.Min(stat.minimum, value)
	stat.maximum = math.Max(stat.maximum, value)
}

func aggregateTree(node *treeNode, summaries map[*treeNode]*treeSummary) *treeSummary {
	summary := &treeSummary{typeAgg: make(map[string]*columnStats)}

	summaries[node] = summary
	accumulateSelf(node, summary)
	accumulateChildren(node, summaries, summary)

	return summary
}

func accumulateSelf(node *treeNode, summary *treeSummary) {
	if node.pkg == nil {
		return
	}

	summary.pkgsTotal = countOne
	summary.typesTotal = len(node.pkg.Types)
	collectPackageStats(node.pkg, summary.typeAgg)
}

func accumulateChildren(
	node *treeNode,
	summaries map[*treeNode]*treeSummary,
	summary *treeSummary,
) {
	// Continue.
	for i := range node.children {
		child := aggregateTree(node.children[i], summaries)

		summary.pkgsTotal += child.pkgsTotal
		summary.typesTotal += child.typesTotal
		mergeStats(summary.typeAgg, child.typeAgg)
	}
}

func branchGlyph(index, total int) string {
	if index == total-countOne {
		return cornerBranch
	}

	return teeBranch
}

func buildTree(report *reusability.Report) *treeNode {
	root := &treeNode{}

	for i := range report.Packages {
		attachPackage(root, &report.Packages[i], report.Module)
	}

	compressChildren(root)

	return root
}

func attachPackage(root *treeNode, pkg *reusability.PackageReport, module string) {
	rel := relPath(pkg.Path, module)

	if rel == pathDot {
		root.pkg = pkg

		return
	}

	node := root

	for seg := range strings.SplitSeq(rel, "/") {
		node = appendTreeChild(&node.children, seg)
	}

	node.pkg = pkg
}

func compressChildren(root *treeNode) {
	for i := range root.children {
		root.children[i].compress()
	}
}

func childIndent(prefix, connector string) string {
	switch connector {
	case teeBranch:
		return prefix + teePad
	case cornerBranch:
		return prefix + cornerPad
	default:
		return prefix
	}
}

func collectPackageStats(pkg *reusability.PackageReport, typeAgg map[string]*columnStats) {
	for i := range pkg.Types {
		reuse := pkg.Types[i].Reusability

		if reuse.Applicable {
			addStat(typeAgg, metrics.MetricReusability, reuse.Value)
		}
	}
}

func formatCell(value float64) string {
	return strconv.FormatFloat(value, 'f', countTwo, floatBits)
}

func meanCell(stat *columnStats, color func(float64) string) tableCell {
	if stat == nil || stat.count == indexZero {
		return naTableCell()
	}

	value := stat.sum / float64(stat.count)

	return tableCell{text: formatCell(value), style: ansiBold + color(value)}
}

func mergeStats(dst, src map[string]*columnStats) {
	for name := range src {
		mergeOneStat(dst, name, src[name])
	}
}

func mergeOneStat(dst map[string]*columnStats, name string, src *columnStats) {
	dstStat := dst[name]

	if dstStat == nil {
		copyStat := *src

		dst[name] = &copyStat

		return
	}

	dstStat.sum += src.sum
	dstStat.count += src.count

	dstStat.minimum = math.Min(dstStat.minimum, src.minimum)
	dstStat.maximum = math.Max(dstStat.maximum, src.maximum)
}

func naTableCell() tableCell {
	return tableCell{text: naCell, style: ansiDim}
}

func nodeTypeAggCells(summary *treeSummary, typeCols []string) []tableCell {
	if summary.typesTotal == indexZero {
		return nil
	}

	cells := make([]tableCell, indexZero, len(typeCols))

	for i := range typeCols {
		name := typeCols[i]
		stat := summary.typeAgg[name]

		cells = append(cells, meanCell(stat, func(value float64) string {
			return valueColor(name, value, stat)
		}))
	}

	return cells
}

func nodeTypeCount(node *treeNode, typeCols []string) int {
	if node.pkg != nil && len(typeCols) > indexZero {
		return len(node.pkg.Types)
	}

	return indexZero
}

func packageNotes(pkg *reusability.PackageReport) []string {
	entries := collectNoteEntries(pkg)

	return formatNoteEntries(pkg, entries)
}

func collectNoteEntries(pkg *reusability.PackageReport) []noteEntry {
	state := &noteCollectState{
		entries: zeroNoteEntries(len(pkg.Types)),
		index:   make(map[string]int),
	}

	for i := range pkg.Types {
		appendNote(state, &pkg.Types[i])
	}

	return state.entries[:state.count]
}

func appendNote(state *noteCollectState, typ *reusability.TypeReport) {
	reuse := typ.Reusability

	if reuse.Reason == emptyString {
		return
	}

	appendNoteType(state, reuse.Reason, typ.Name)
}

func appendNoteType(state *noteCollectState, reason, typeName string) {
	pos, ok := state.index[reason]

	if !ok {
		pos = state.count
		state.index[reason] = pos
		state.entries[pos] = noteEntry{reason: reason}
		state.count++
	}

	state.entries[pos].types = append(state.entries[pos].types, typeName)
}

func formatNoteEntries(pkg *reusability.PackageReport, entries []noteEntry) []string {
	notes := make([]string, indexZero, len(entries))

	for i := range entries {
		notes = append(notes, formatOneNote(pkg, &entries[i]))
	}

	return notes
}

func formatOneNote(pkg *reusability.PackageReport, entry *noteEntry) string {
	who := strings.Join(entry.types, ", ")

	if len(entry.types) == len(pkg.Types) && len(pkg.Types) > countOne {
		who = "all types"
	}

	return metrics.MetricReusability + ": " + entry.reason + " (" + who + ")"
}

func paint(args *paintArgs) string {
	if !args.color || args.style == emptyString {
		return args.text
	}

	return args.style + args.text + ansiReset
}

func relPath(path, module string) string {
	if module == emptyString {
		return path
	}

	if path == module {
		return pathDot
	}

	if strings.HasPrefix(path, module+"/") {
		return path[len(module)+countOne:]
	}

	return path
}

func reportColumns(report *reusability.Report) []string {
	for i := range report.Packages {
		if len(report.Packages[i].Types) > indexZero {
			return []string{metrics.MetricReusability}
		}
	}

	return nil
}

func reusabilityCell(
	result *metrics.MetricResult,
	typeCols []string,
	stats map[string]*columnStats,
) []tableCell {
	// Continue.
	cells := make([]tableCell, indexZero, len(typeCols))

	for i := range typeCols {
		cells = append(cells, oneReusabilityCell(result, typeCols[i], stats))
	}

	return cells
}

func oneReusabilityCell(
	result *metrics.MetricResult,
	name string,
	stats map[string]*columnStats,
) tableCell {
	// Continue.
	if name != metrics.MetricReusability || !result.Applicable {
		return naTableCell()
	}

	return tableCell{
		text:  formatCell(result.Value),
		style: valueColor(name, result.Value, stats[name]),
	}
}

func (cell *tableCell) width() int {
	return utf8.RuneCountInString(cell.prefix) + utf8.RuneCountInString(cell.text)
}

func emitModuleSummary(table *textTable, node *treeNode, summary *treeSummary) {
	nodeRow(table, node, summary, emptyString)

	typeCount := nodeTypeCount(node, table.typeCols)

	for i := range typeCount {
		typeRow(
			table, node, summary, i, emptyString, branchGlyph(i, typeCount),
		)
	}
}

func emitNode(
	table *textTable,
	node *treeNode,
	summaries map[*treeNode]*treeSummary,
	prefix, connector string,
) {
	nodeRow(table, node, summaries[node], prefix+connector)

	childPrefix := childIndent(prefix, connector)
	typeCount := nodeTypeCount(node, table.typeCols)
	total := typeCount + len(node.children)
	emitTypeBranches(table, node, summaries[node], childPrefix, typeCount, total)
	emitChildNodes(table, node, summaries, childPrefix, typeCount, total)
}

func emitTypeBranches(
	table *textTable,
	node *treeNode,
	summary *treeSummary,
	childPrefix string,
	typeCount, total int,
) {
	for i := range typeCount {
		typeRow(table, node, summary, i, childPrefix, branchGlyph(i, total))
	}
}

func emitChildNodes(
	table *textTable,
	node *treeNode,
	summaries map[*treeNode]*treeSummary,
	childPrefix string,
	typeCount, total int,
) {
	for i := range node.children {
		if typeCount+i > indexZero {
			table.rows = append(table.rows, []tableCell{{prefix: childPrefix + "│"}})
		}

		emitNode(
			table, node.children[i], summaries,
			childPrefix, branchGlyph(typeCount+i, total),
		)
	}
}

func nodeRow(table *textTable, node *treeNode, summary *treeSummary, label string) {
	row := make([]tableCell, indexZero, len(table.typeCols)+1)

	row = append(row, tableCell{prefix: label, text: node.name, style: ansiBold})
	row = append(row, nodeTypeAggCells(summary, table.typeCols)...)
	table.rows = append(table.rows, row)
}

func typeRow(
	table *textTable,
	node *treeNode,
	summary *treeSummary,
	index int,
	prefix, connector string,
) {
	typ := &node.pkg.Types[index]
	row := make([]tableCell, indexZero, len(table.typeCols)+1)

	row = append(row, tableCell{prefix: prefix + connector, text: typ.Name})
	row = append(row, reusabilityCell(&typ.Reusability, table.typeCols, summary.typeAgg)...)
	table.rows = append(table.rows, row)
}

func thresholdColor(bias scoreBias, score float64) string {
	if bias == biasLowerBetter {
		score = float64(countOne) - score
	}

	switch {
	case score >= qualityHigh:
		return ansiGreen
	case score >= qualityMedium:
		return ansiYellow
	default:
		return ansiRed
	}
}

func appendTreeChild(children *[]*treeNode, name string) *treeNode {
	for i := range *children {
		if (*children)[i].name == name {
			return (*children)[i]
		}
	}

	child := &treeNode{name: name}

	*children = append(*children, child)

	return child
}

func (node *treeNode) compress() {
	for node.pkg == nil && len(node.children) == countOne {
		node.absorbOnlyChild()
	}

	compressTreeChildren(node.children)
}

func compressTreeChildren(children []*treeNode) {
	for i := range children {
		children[i].compress()
	}
}

func (node *treeNode) absorbOnlyChild() {
	child := node.children[indexZero]

	node.name = node.name + "/" + child.name
	node.pkg = child.pkg
	node.children = child.children
}

func valueColor(name string, value float64, stat *columnStats) string {
	quality, ok := qualityForMetric(name)

	if !ok {
		return emptyString
	}

	if quality.bounded {
		return thresholdColor(quality.bias, value)
	}

	normalized, ok := tryNormalize(value, stat)

	if !ok {
		return emptyString
	}

	return thresholdColor(quality.bias, normalized)
}

func tryNormalize(value float64, stat *columnStats) (float64, bool) {
	if stat == nil || stat.maximum == stat.minimum {
		return indexZero, false
	}

	return (value - stat.minimum) / (stat.maximum - stat.minimum), true
}

func writeNotes(builder *strings.Builder, report *reusability.Report, color bool) error {
	writer := &notesWriter{builder: builder, color: color}

	for i := range report.Packages {
		err := writeNotesPackage(writer, &report.Packages[i])
		if err != nil {
			return fmt.Errorf("write package notes: %w", err)
		}
	}

	return nil
}

func writeNotesPackage(writer *notesWriter, pkg *reusability.PackageReport) error {
	notes := packageNotes(pkg)

	if len(notes) == indexZero {
		return nil
	}

	err := ensureNotesHeader(writer)
	if err != nil {
		return fmt.Errorf("write notes header: %w", err)
	}

	err = writeNotesBody(&notesBodyArgs{
		builder: writer.builder, path: pkg.Path, notes: notes, color: writer.color,
	})
	if err != nil {
		return fmt.Errorf("write notes body: %w", err)
	}

	return nil
}

func ensureNotesHeader(writer *notesWriter) error {
	if writer.wrote {
		return nil
	}

	err := writeNotesHeader(writer.builder, writer.color)
	if err != nil {
		return fmt.Errorf("write header lines: %w", err)
	}

	writer.wrote = true

	return nil
}

func writeNotesHeader(builder *strings.Builder, color bool) error {
	err := writeBuilderStrings(
		builder,
		newline,
		paint(&paintArgs{text: notesLabel, style: ansiDim, color: color}),
		newline,
	)
	if err != nil {
		return fmt.Errorf("write notes title: %w", err)
	}

	return nil
}

func writeNotesBody(args *notesBodyArgs) error {
	err := writeNotesPath(args)
	if err != nil {
		return fmt.Errorf("write notes path: %w", err)
	}

	err = writeNoteLines(args)
	if err != nil {
		return fmt.Errorf("write note lines: %w", err)
	}

	return nil
}

func writeNotesPath(args *notesBodyArgs) error {
	err := writeBuilderStrings(
		args.builder,
		"  ",
		paint(&paintArgs{text: args.path, style: ansiDim, color: args.color}),
		newline,
	)
	if err != nil {
		return fmt.Errorf("write path strings: %w", err)
	}

	return nil
}

func writeNoteLines(args *notesBodyArgs) error {
	for i := range args.notes {
		err := writeBuilderStrings(
			args.builder,
			"    ",
			paint(&paintArgs{text: args.notes[i], style: ansiDim, color: args.color}),
			newline,
		)
		if err != nil {
			return fmt.Errorf("write note line: %w", err)
		}
	}

	return nil
}
