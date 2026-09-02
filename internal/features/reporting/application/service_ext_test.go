package application_test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"strings"
	"testing"

	reporting "github.com/gostafa/reusability/internal/features/reporting/application"
	"github.com/gostafa/reusability/internal/features/reporting/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/reusability"
)

type bufSink struct{ buf *bytes.Buffer }

func (b bufSink) Open() (io.WriteCloser, error) { return nopCloser{b.buf}, nil }

type nopCloser struct{ io.Writer }

func (nopCloser) Close() error { return nil }

func report() reusability.Report {
	return reusability.Report{
		SchemaVersion: "6",
		Tool:          reusability.ToolInfo{Name: "reusability", Version: "test"},
		Module:        "example.com/m",
		Packages: []reusability.PackageReport{
			{
				Path: "example.com/m/a",
				Types: []reusability.TypeReport{{
					Name: "A",
					Reusability: metrics.MetricResult{
						Name:       "reusability",
						Scope:      metrics.ScopeType,
						Value:      0.7,
						Applicable: true,
						Definition: "d",
					},
				}},
			},
		},
	}
}

// Black-box: the text format includes the module and the type row.
func TestWriteText(t *testing.T) {
	t.Parallel()

	sink := bufSink{&bytes.Buffer{}}
	err := reporting.Write(report(), domain.FormatText, sink, domain.TextOptions{})
	if err != nil {
		t.Fatal(err)
	}

	out := sink.buf.String()
	if !strings.Contains(out, "example.com/m") || !strings.Contains(out, "A") {
		t.Fatalf("text output missing content:\n%s", out)
	}
}

// Black-box: the JSON format is valid and versioned.
func TestWriteJSON(t *testing.T) {
	t.Parallel()

	sink := bufSink{&bytes.Buffer{}}
	err := reporting.Write(report(), domain.FormatJSON, sink, domain.TextOptions{})
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	err = json.Unmarshal(sink.buf.Bytes(), &got)
	if err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if got["schema_version"] != "6" {
		t.Errorf("schema_version = %v", got["schema_version"])
	}
	pkg := got["packages"].([]any)[0].(map[string]any)
	if _, ok := pkg["afferent"]; ok {
		t.Fatal("package afferent should not appear in JSON schema v6")
	}
	typ := pkg["types"].([]any)[0].(map[string]any)
	if typ["name"] != "A" {
		t.Fatalf("type name = %+v", typ["name"])
	}
	reuse := typ["reusability"].(map[string]any)
	if reuse["value"].(float64) != 0.7 {
		t.Fatalf("reusability = %+v", reuse)
	}
}

// Black-box: the CSV format starts with the canonical header and has a row per
// type.
func TestWriteCSV(t *testing.T) {
	t.Parallel()

	sink := bufSink{&bytes.Buffer{}}
	if err := reporting.Write(report(), domain.FormatCSV, sink, domain.TextOptions{}); err != nil {
		t.Fatal(err)
	}

	records, err := csv.NewReader(sink.buf).ReadAll()
	if err != nil {
		t.Fatalf("invalid CSV: %v", err)
	}

	if len(records) < 2 {
		t.Fatalf("csv has %d rows, want header + data", len(records))
	}

	header := strings.Join(records[0], ",")
	if header != strings.Join(domain.CSVHeader(), ",") {
		t.Errorf("csv header = %q", header)
	}
}
