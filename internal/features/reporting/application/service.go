package application

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gostafa/reusability/internal/features/reporting/domain"
	"github.com/gostafa/reusability/internal/features/reporting/ports/outbound"
	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/reusability"
)

// jsonMarshal is a seam so tests can force encoding failures.
var jsonMarshal = json.Marshal

// Write renders the report in the given format into the sink. Options are
// read only by the text format.
func Write(
	report reusability.Report,
	format domain.Format,
	sink outbound.Sink,
	opts domain.TextOptions,
) error {
	w, err := sink.Open()
	if err != nil {
		return err
	}

	renderErr := render(w, report, format, opts)
	if renderErr != nil {
		_ = w.Close()

		return renderErr
	}

	return w.Close()
}

func render(
	w io.Writer,
	report reusability.Report,
	format domain.Format,
	opts domain.TextOptions,
) error {
	switch format {
	case domain.FormatText:
		_, err := io.WriteString(w, domain.Text(report, opts))

		return err
	case domain.FormatJSON:
		return renderJSON(w, report)
	case domain.FormatCSV:
		// WriteAll flushes; a separate header Write cannot surface bufio
		// errors until Flush, so header and records go through one call.
		rows := append([][]string{domain.CSVHeader()}, domain.CSVRecords(report)...)

		return csv.NewWriter(w).WriteAll(rows)
	case domain.FormatWeb:
		return renderWeb(w, report)
	default:
		return fmt.Errorf("unknown report format %q", format)
	}
}

// jsonReport mirrors the versioned report schema (§ output). Metric maps
// are orderedMetrics so keys always appear in the fixed metric order.
type jsonReport struct {
	// SchemaVersion is the report schema version.
	SchemaVersion string `json:"schema_version"`
	// Tool identifies the producing tool.
	Tool jsonTool `json:"tool"`
	// Packages are the analyzed packages in report order.
	Packages []jsonPackage `json:"packages"`
}

// String summarizes the report envelope for debugging.
func (r jsonReport) String() string {
	return fmt.Sprintf("schema %s, tool %v, %d packages", r.SchemaVersion, r.Tool, len(r.Packages))
}

type jsonTool struct {
	// Name is the tool's canonical name.
	Name string `json:"name"`
	// Version is the tool's build version.
	Version string `json:"version"`
}

type jsonPackage struct {
	// Path is the package's import path.
	Path string `json:"path"`
	// Afferent counts analyzed packages importing this package (Ca).
	Afferent int `json:"afferent"`
	// Efferent counts this package's in-scope imports (Ce).
	Efferent int `json:"efferent"`
	// Funcs counts the package's declared functions and methods.
	Funcs int `json:"funcs"`
	// Vars counts top-level variable names declared in the package.
	Vars int `json:"vars"`
	// Consts counts top-level constant names declared in the package.
	Consts int `json:"consts"`
	// Variables are top-level variable declarations.
	Variables []jsonDeclaration `json:"variables"`
	// Constants are top-level constant declarations.
	Constants []jsonDeclaration `json:"constants"`
	// Functions are top-level function declarations, excluding methods.
	Functions []jsonFunction `json:"functions"`
	// Types are the package's analyzed types in report order.
	Types []jsonType `json:"types"`
}

// String summarizes one package entry for debugging.
func (p jsonPackage) String() string {
	return fmt.Sprintf("%s: %d types", p.Path, len(p.Types))
}

type jsonType struct {
	// Name is the type's declared name.
	Name string `json:"name"`
	// Exported reports whether the type name is exported.
	Exported bool `json:"exported"`
	// Kind classifies the type's underlying type.
	Kind string `json:"kind"`
	// Position locates the type declaration in source.
	Position jsonPosition `json:"position"`
	// Fields is the struct field count (embedded fields count one).
	Fields int `json:"fields"`
	// FieldDetails holds the struct fields in declaration order.
	FieldDetails []jsonField `json:"field_details"`
	// Methods is the declared method count.
	Methods int `json:"methods"`
	// MethodDetails holds the declared methods in report order.
	MethodDetails []jsonFunction `json:"method_details"`
	// Metrics maps metric names to results in the fixed order.
	Metrics orderedMetrics `json:"metrics"`
}

type jsonPosition struct {
	// File is the declaration source file.
	File string `json:"file"`
	// Line is the 1-based source line.
	Line int `json:"line"`
	// Column is the 1-based source column.
	Column int `json:"column"`
}

type jsonDeclaration struct {
	// Name is the declared identifier.
	Name string `json:"name"`
	// Exported reports whether the declaration name is exported.
	Exported bool `json:"exported"`
	// Position locates the declaration in source.
	Position jsonPosition `json:"position"`
}

type jsonField struct {
	// Name is the field name.
	Name string `json:"name"`
	// Exported reports whether the field name is exported.
	Exported bool `json:"exported"`
	// Embedded marks an embedded field.
	Embedded bool `json:"embedded"`
}

type jsonFunction struct {
	// Name is the declared function or method name.
	Name string `json:"name"`
	// Exported reports whether the function name is exported.
	Exported bool `json:"exported"`
	// Receiver is the declaring type name for methods; empty for free funcs.
	Receiver string `json:"receiver"`
	// Position locates the declaration in source.
	Position jsonPosition `json:"position"`
	// Lines is the inclusive source line count from func keyword to end.
	Lines int `json:"lines"`
	// Cyclomatic is the cyclomatic complexity score.
	Cyclomatic int `json:"cyclomatic"`
	// Branches are the syntax counts feeding cyclomatic complexity.
	Branches jsonBranchStats `json:"branches"`
}

type jsonBranchStats struct {
	Ifs         int `json:"ifs"`
	Fors        int `json:"fors"`
	Ranges      int `json:"ranges"`
	Cases       int `json:"cases"`
	SelectComms int `json:"select_comms"`
	LogicalOps  int `json:"logical_ops"`
}

// jsonMetric serializes one MetricResult. A non-applicable metric carries
// its reason and no value — never a fake zero.
type jsonMetric struct {
	// Scope is the kind of entity the metric describes.
	Scope string `json:"scope"`
	// Value is the metric value, present only when applicable.
	Value *float64 `json:"value,omitempty"`
	// Applicable reports whether the value may be read.
	Applicable bool `json:"applicable"`
	// Reason explains non-applicability or dropped components.
	Reason string `json:"reason,omitempty"`
	// Definition is the versioned formula identifier.
	Definition string `json:"definition"`
}

// orderedMetrics marshals as a JSON object keyed by metric name, preserving
// slice order (the fixed metric order).
type orderedMetrics []metrics.MetricResult

// MarshalJSON writes the object with keys in the fixed metric order.
func (m orderedMetrics) MarshalJSON() ([]byte, error) {
	return encodeOrderedMetrics(m)
}

// encodeOrderedMetrics assembles the ordered JSON object one name→metric
// pair at a time.
func encodeOrderedMetrics(results []metrics.MetricResult) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')

	for i, r := range results {
		if i > 0 {
			buf.WriteByte(',')
		}

		err := encodeMetricEntry(&buf, r)
		if err != nil {
			return nil, err
		}
	}

	buf.WriteByte('}')

	return buf.Bytes(), nil
}

// encodeMetricEntry writes one name→metric pair. A non-applicable metric
// carries its reason and no value — never a fake zero.
func encodeMetricEntry(buf *bytes.Buffer, r metrics.MetricResult) error {
	key, err := jsonMarshal(r.Name)
	if err != nil {
		return err
	}

	buf.Write(key)
	buf.WriteByte(':')

	out := jsonMetric{
		Scope:      string(r.Scope),
		Applicable: r.Applicable,
		Reason:     r.Reason,
		Definition: r.Definition,
	}
	if r.Applicable {
		value := r.Value
		out.Value = &value
	}

	encoded, err := jsonMarshal(out)
	if err != nil {
		return err
	}

	buf.Write(encoded)

	return nil
}

// buildJSONReport maps the report onto the versioned JSON schema. It is
// shared by the JSON format and the web report's embedded payload.
func buildJSONReport(report reusability.Report) jsonReport {
	out := jsonReport{
		SchemaVersion: report.SchemaVersion,
		Tool:          jsonTool{Name: report.Tool.Name, Version: report.Tool.Version},
		Packages:      make([]jsonPackage, len(report.Packages)),
	}
	for i, pkg := range report.Packages {
		jp := jsonPackage{
			Path:      pkg.Path,
			Afferent:  pkg.Afferent,
			Efferent:  pkg.Efferent,
			Funcs:     pkg.ExportedFuncs + pkg.UnexportedFuncs,
			Vars:      pkg.Vars,
			Consts:    pkg.Consts,
			Variables: declarations(pkg.Variables),
			Constants: declarations(pkg.Constants),
			Functions: functions(pkg.Functions),

			Types: make([]jsonType, len(pkg.Types)),
		}
		for j, t := range pkg.Types {
			jp.Types[j] = jsonType{
				Name:          t.Name,
				Exported:      t.Exported,
				Kind:          t.Kind,
				Position:      position(t.Position),
				Fields:        t.Fields,
				FieldDetails:  fields(t.FieldDetails),
				Methods:       t.Methods,
				MethodDetails: functions(t.MethodDetails),
				Metrics:       orderedMetrics(t.Metrics),
			}
		}

		out.Packages[i] = jp
	}

	return out
}

func position(pos reusability.Position) jsonPosition {
	return jsonPosition{File: pos.File, Line: pos.Line, Column: pos.Column}
}

func declarations(decls []reusability.DeclarationReport) []jsonDeclaration {
	out := make([]jsonDeclaration, len(decls))
	for i, d := range decls {
		out[i] = jsonDeclaration{Name: d.Name, Exported: d.Exported, Position: position(d.Position)}
	}

	return out
}

func fields(fieldDetails []reusability.FieldReport) []jsonField {
	out := make([]jsonField, len(fieldDetails))
	for i, f := range fieldDetails {
		out[i] = jsonField{Name: f.Name, Exported: f.Exported, Embedded: f.Embedded}
	}

	return out
}

func functions(functionDetails []reusability.FunctionReport) []jsonFunction {
	out := make([]jsonFunction, len(functionDetails))
	for i, fn := range functionDetails {
		out[i] = jsonFunction{
			Name:       fn.Name,
			Exported:   fn.Exported,
			Receiver:   fn.Receiver,
			Position:   position(fn.Position),
			Lines:      fn.Lines,
			Cyclomatic: fn.Cyclomatic,
			Branches:   branches(fn.Branches),
		}
	}

	return out
}

func branches(stats reusability.BranchStats) jsonBranchStats {
	return jsonBranchStats{
		Ifs:         stats.Ifs,
		Fors:        stats.Fors,
		Ranges:      stats.Ranges,
		Cases:       stats.Cases,
		SelectComms: stats.SelectComms,
		LogicalOps:  stats.LogicalOps,
	}
}

func renderJSON(w io.Writer, report reusability.Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	return enc.Encode(buildJSONReport(report))
}
