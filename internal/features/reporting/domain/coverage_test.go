package domain

import (
	"strings"
	"testing"

	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/reusability"
)

func TestRelPathEdges(t *testing.T) {
	if got := relPath("example.com/m/p", ""); got != "example.com/m/p" {
		t.Fatalf("empty module: got %q", got)
	}
	if got := relPath("other.com/x", "example.com/m"); got != "other.com/x" {
		t.Fatalf("outside module: got %q", got)
	}
}

func TestTextEmptyPackages(t *testing.T) {
	got := Text(reusability.Report{
		SchemaVersion: "1",
		Tool:          reusability.ToolInfo{Name: "reusability", Version: "test"},
		Module:        "example.com/m",
	}, TextOptions{})
	if !strings.Contains(got, "module example.com/m") || strings.Contains(got, "PATH / TYPE") {
		t.Fatalf("empty packages output unexpected:\n%s", got)
	}
}

func TestTextMultiSectionSpacerAndMissingMetrics(t *testing.T) {
	report := reusability.Report{
		SchemaVersion: "1",
		Tool:          reusability.ToolInfo{Name: "reusability", Version: "test"},
		Module:        "example.com/m",
		Packages: []reusability.PackageReport{
			{
				Path: "example.com/m",
			},
			{
				Path: "example.com/m/leaf",
				Types: []reusability.TypeReport{{
					Name: "T",
					Reusability: typeMetric(metrics.MetricReusability, 0.5),
				}},
			},
		},
	}

	got := Text(report, TextOptions{Color: true, Explain: true})
	if !strings.Contains(got, "\n\n") {
		t.Fatalf("expected blank spacer between sections:\n%s", got)
	}
	if strings.Contains(got, ansiGreen+"9.00") || strings.Contains(got, ansiRed+"9.00") {
		t.Fatalf("unknown metric was quality-colored:\n%q", got)
	}
}

func TestTextExplainAllTypesAndSkipEmptyNotes(t *testing.T) {
	report := reusability.Report{
		SchemaVersion: "1",
		Tool:          reusability.ToolInfo{Name: "reusability", Version: "test"},
		Module:        "example.com/m",
		Packages: []reusability.PackageReport{
			{
				Path: "example.com/m/quiet",
			},
			{
				Path: "example.com/m/noisy",
				Types: []reusability.TypeReport{
					{Name: "A", Reusability: metrics.MetricResult{
						Name:       metrics.MetricReusability,
						Scope:      metrics.ScopeType,
						Applicable: false,
						Reason:     "every component dropped",
					}},
					{Name: "B", Reusability: metrics.MetricResult{
						Name:       metrics.MetricReusability,
						Scope:      metrics.ScopeType,
						Applicable: false,
						Reason:     "every component dropped",
					}},
				},
			},
		},
	}

	got := Text(report, TextOptions{Explain: true})
	if !strings.Contains(got, "reusability: every component dropped (all types)") {
		t.Fatalf("want all-types aggregation:\n%s", got)
	}
	notesIdx := strings.Index(got, "\nnotes\n")
	if notesIdx < 0 {
		t.Fatalf("notes section missing:\n%s", got)
	}
	if strings.Contains(got[notesIdx:], "quiet") {
		t.Fatalf("quiet package should be skipped in notes:\n%s", got[notesIdx:])
	}
}

func TestValueColorUnknownMetric(t *testing.T) {
	if got := valueColor("not-a-metric", 1, &columnStats{min: 0, max: 2, count: 2}); got != "" {
		t.Fatalf("valueColor = %q, want empty", got)
	}
}

func TestMeanCellNilStats(t *testing.T) {
	cell := meanCell(nil, func(float64) string { return "" })
	if cell.text != naCell {
		t.Fatalf("meanCell(nil) = %q, want %q", cell.text, naCell)
	}
}
