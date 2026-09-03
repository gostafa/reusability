// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gostafa/reusability/internal/features/reporting/domain"
	"github.com/gostafa/reusability/internal/features/reporting/ports/outbound"
	"github.com/gostafa/reusability/reusability"
)

// Write renders the report in the given format into the sink. Options are
// read only by the text format.
func Write(req *WriteRequest) error {
	writer, err := outbound.Open(req.Sink)
	if err != nil {
		return fmt.Errorf("open sink: %w", err)
	}

	renderErr := render(&renderInput{
		writer: writer, report: &req.Report, format: req.Format, opts: req.Options,
	})
	closeErr := writer.Close()

	if renderErr != nil {
		return fmt.Errorf("render report: %w", renderErr)
	}

	if closeErr != nil {
		return fmt.Errorf("close sink: %w", closeErr)
	}

	return nil
}

// WriteDocs renders the metrics guide into the sink. It needs no report:
// the page documents the tool, not one run.
func WriteDocs(sink outbound.Sink, toolVersion string) error {
	writer, err := outbound.Open(sink)
	if err != nil {
		return fmt.Errorf("open docs sink: %w", err)
	}

	renderErr := renderDocs(writer, toolVersion)
	closeErr := writer.Close()

	if renderErr != nil {
		return fmt.Errorf("render docs: %w", renderErr)
	}

	if closeErr != nil {
		return fmt.Errorf("close docs sink: %w", closeErr)
	}

	return nil
}

// String summarizes one package entry for debugging.
func (pkg jsonPackage) String() string {
	return fmt.Sprintf("%s: %d types", pkg.Path, len(pkg.Types))
}

// jsonReportString summarizes the report envelope for debugging.
func jsonReportString(report *jsonReport) string {
	return fmt.Sprintf(
		"schema %s, tool %v, %d packages",
		report.SchemaVersion, report.Tool, len(report.Packages),
	)
}

func buildJSONReport(report *reusability.Report) jsonReport {
	out := jsonReport{
		SchemaVersion: report.SchemaVersion,
		Tool:          jsonTool{Name: report.Tool.Name, Version: report.Tool.Version},
		Packages:      make([]jsonPackage, len(report.Packages)),
	}

	fillJSONPackages(out.Packages, report.Packages)

	return out
}

func fillJSONPackages(dst []jsonPackage, pkgs []reusability.PackageReport) {
	for i := range pkgs {
		dst[i] = jsonPackageFrom(&pkgs[i])
	}
}

func jsonPackageFrom(pkg *reusability.PackageReport) jsonPackage {
	out := jsonPackage{
		Path:  pkg.Path,
		Types: make([]jsonType, len(pkg.Types)),
	}

	for idx := range pkg.Types {
		out.Types[idx] = jsonType{
			Name:        pkg.Types[idx].Name,
			Reusability: metricJSON(&pkg.Types[idx].Reusability),
		}
	}

	return out
}

func marshalDocsWith(toolVersion string, marshal func(any) ([]byte, error)) ([]byte, error) {
	entries := domain.MetricDocs()
	out := docsPayload{
		Tool: jsonTool{Name: reusability.ToolName(), Version: toolVersion},
		Docs: make([]jsonMetricDoc, len(entries)),
	}

	for i := range entries {
		out.Docs[i] = jsonMetricDocFrom(&entries[i])
	}

	payload, err := marshal(out)
	if err != nil {
		return nil, fmt.Errorf("marshal docs: %w", err)
	}

	return payload, nil
}

func jsonMetricDocFrom(doc *domain.MetricDoc) jsonMetricDoc {
	return jsonMetricDoc{
		Name:           doc.Name,
		Label:          doc.Label,
		FullName:       doc.FullName,
		Scope:          string(doc.Scope),
		Definition:     doc.Definition,
		FormulaMathML:  doc.FormulaMathML,
		FormulaLaTeX:   doc.FormulaLaTeX,
		Summary:        doc.Summary,
		How:            doc.HowCalculated,
		Interpretation: doc.Interpretation,
		NotApplicable:  doc.NotApplicable,
		Direction:      doc.Direction,
		Bounded:        doc.Bounded,
		Example:        doc.Example,
	}
}

func metricJSON(metric *reusability.MetricResult) jsonMetric {
	out := jsonMetric{
		Applicable: metric.Applicable,
		Reason:     metric.Reason,
		Definition: metric.Definition,
	}

	if metric.Applicable {
		value := metric.Value

		out.Value = &value
	}

	return out
}

func render(input *renderInput) error {
	err := renderStructured(input)
	if err != nil {
		return fmt.Errorf("render: %w", err)
	}

	return nil
}

func renderStructured(input *renderInput) error {
	renderer, ok := rendererFor(input)

	if !ok {
		return fmt.Errorf("%w: %q", errUnknownFormat, input.format)
	}

	err := renderer()
	if err != nil {
		return fmt.Errorf("render %s: %w", input.format, err)
	}

	return nil
}

func rendererFor(input *renderInput) (func() error, bool) {
	renderers := map[domain.Format]func() error{
		domain.FormatText: func() error { return renderText(input.writer, input.report, input.opts) },
		domain.FormatJSON: func() error { return renderJSON(input.writer, input.report) },
		domain.FormatCSV:  func() error { return renderCSV(input.writer, input.report) },
		domain.FormatWeb:  func() error { return renderWeb(input.writer, input.report) },
	}

	renderer, ok := renderers[input.format]

	return renderer, ok
}

func renderText(writer io.Writer, report *reusability.Report, opts domain.TextOptions) error {
	text, err := domain.Text(report, opts)
	if err != nil {
		return fmt.Errorf("build text: %w", err)
	}

	err = writeFullString(writer, text, "write text")
	if err != nil {
		return fmt.Errorf("write text body: %w", err)
	}

	return nil
}

func renderCSV(writer io.Writer, report *reusability.Report) error {
	rows := append([][]string{domain.CSVHeader()}, domain.CSVRecords(report)...)

	err := csv.NewWriter(writer).WriteAll(rows)
	if err != nil {
		return fmt.Errorf("write csv: %w", err)
	}

	return nil
}

func renderDocs(writer io.Writer, toolVersion string) error {
	err := renderDocsWith(writer, toolVersion, json.Marshal)
	if err != nil {
		return fmt.Errorf("render docs with: %w", err)
	}

	return nil
}

func renderDocsWith(
	writer io.Writer,
	toolVersion string,
	marshal func(any) ([]byte, error),
) error {
	// Continue.
	page, err := buildDocsPage(toolVersion, marshal)
	if err != nil {
		return fmt.Errorf("build docs page: %w", err)
	}

	err = writeFullString(writer, page, "write docs")
	if err != nil {
		return fmt.Errorf("write docs body: %w", err)
	}

	return nil
}

func buildDocsPage(toolVersion string, marshal func(any) ([]byte, error)) (string, error) {
	payload, err := marshalDocsWith(toolVersion, marshal)
	if err != nil {
		return emptyString, fmt.Errorf("marshal docs payload: %w", err)
	}

	page, err := injectPlaceholder(docsTemplate, docsDataPlaceholder, string(payload))
	if err != nil {
		return emptyString, fmt.Errorf("inject docs data: %w", err)
	}

	return page, nil
}

func renderJSON(writer io.Writer, report *reusability.Report) error {
	enc := json.NewEncoder(writer)
	enc.SetIndent(emptyString, jsonIndent)

	err := enc.Encode(buildJSONReport(report))
	if err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	return nil
}

func renderWeb(writer io.Writer, report *reusability.Report) error {
	err := renderWebWith(writer, report, json.Marshal)
	if err != nil {
		return fmt.Errorf("render web with: %w", err)
	}

	return nil
}

func renderWebWith(
	writer io.Writer,
	report *reusability.Report,
	marshal func(any) ([]byte, error),
) error {
	// Continue.
	page, err := buildWebPageWith(report, marshal)
	if err != nil {
		return fmt.Errorf("build web page: %w", err)
	}

	err = writeFullString(writer, page, "write web")
	if err != nil {
		return fmt.Errorf("write web body: %w", err)
	}

	return nil
}

func buildWebPageWith(
	report *reusability.Report,
	marshal func(any) ([]byte, error),
) (string, error) {
	// Continue.
	payload, err := marshalWebReport(report, marshal)
	if err != nil {
		return emptyString, fmt.Errorf("marshal web report: %w", err)
	}

	docs, err := marshalDocsWith(report.Tool.Version, marshal)
	if err != nil {
		return emptyString, fmt.Errorf("marshal web docs: %w", err)
	}

	page, err := injectWebPayloads(string(docs), string(payload))
	if err != nil {
		return emptyString, fmt.Errorf("inject web payloads: %w", err)
	}

	return page, nil
}

func marshalWebReport(
	report *reusability.Report,
	marshal func(any) ([]byte, error),
) ([]byte, error) {
	// Continue.
	payload, err := marshal(webPayload{
		Module: report.Module,
		Report: buildJSONReport(report),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal web payload: %w", err)
	}

	return payload, nil
}

func injectWebPayloads(docs, payload string) (string, error) {
	page, err := injectPlaceholder(webTemplate, docsDataPlaceholder, docs)
	if err != nil {
		return emptyString, fmt.Errorf("inject docs placeholder: %w", err)
	}

	page, err = injectPlaceholder(page, webDataPlaceholder, payload)
	if err != nil {
		return emptyString, fmt.Errorf("inject web placeholder: %w", err)
	}

	return page, nil
}

func injectPlaceholder(template, placeholder, value string) (string, error) {
	if !strings.Contains(template, placeholder) {
		return emptyString, fmt.Errorf("%w: %s", errMissingPlaceholder, placeholder)
	}

	return strings.Replace(template, placeholder, value, replaceOnce), nil
}

func writeFullString(writer io.Writer, text, operation string) error {
	written, err := io.WriteString(writer, text)
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}

	if written != len(text) {
		return fmt.Errorf("%s: wrote %d of %d: %w", operation, written, len(text), errShortWrite)
	}

	return nil
}
