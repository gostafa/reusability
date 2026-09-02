package application

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gostafa/reusability/internal/features/reporting/domain"
	"github.com/gostafa/reusability/internal/features/reporting/ports/outbound"
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

// jsonReport mirrors the versioned report schema (§ output).
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
	// Reusability is the type-level reusability index result.
	Reusability jsonMetric `json:"reusability"`
}

// jsonMetric serializes one MetricResult. A non-applicable metric carries
// its reason and no value — never a fake zero.
type jsonMetric struct {
	// Value is the metric value, present only when applicable.
	Value *float64 `json:"value,omitempty"`
	// Applicable reports whether the value may be read.
	Applicable bool `json:"applicable"`
	// Reason explains non-applicability or dropped components.
	Reason string `json:"reason,omitempty"`
	// Definition is the versioned formula identifier.
	Definition string `json:"definition"`
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
			Path:  pkg.Path,
			Types: make([]jsonType, len(pkg.Types)),
		}
		for j, t := range pkg.Types {
			jp.Types[j] = jsonType{
				Name:          t.Name,
				Reusability:   metricJSON(t.Reusability),
			}
		}

		out.Packages[i] = jp
	}

	return out
}

func metricJSON(r reusability.MetricResult) jsonMetric {
	out := jsonMetric{
		Applicable: r.Applicable,
		Reason:     r.Reason,
		Definition: r.Definition,
	}
	if r.Applicable {
		value := r.Value
		out.Value = &value
	}

	return out
}

func renderJSON(w io.Writer, report reusability.Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	return enc.Encode(buildJSONReport(report))
}
