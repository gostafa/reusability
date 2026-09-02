// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"regexp"
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
	got, err := Text(&reusability.Report{
		SchemaVersion: "1",
		Tool:          reusability.ToolInfo{Name: reusability.ToolName(), Version: "test"},
		Module:        "example.com/m",
	}, TextOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "module example.com/m") || strings.Contains(got, pathTypeHeader) {
		t.Fatalf("empty packages output unexpected:\n%s", got)
	}
}

func TestTextMultiSectionSpacerAndMissingMetrics(t *testing.T) {
	report := reusability.Report{
		SchemaVersion: "1",
		Tool:          reusability.ToolInfo{Name: reusability.ToolName(), Version: "test"},
		Module:        "example.com/m",
		Packages: []reusability.PackageReport{
			{
				Path: "example.com/m",
			},
			{
				Path: "example.com/m/leaf",
				Types: []reusability.TypeReport{{
					Name:        "T",
					Reusability: typeMetric(metrics.MetricReusability, 0.5),
				}},
			},
		},
	}

	got, err := Text(&report, TextOptions{Color: true, Explain: true})
	if err != nil {
		t.Fatal(err)
	}
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
		Tool:          reusability.ToolInfo{Name: reusability.ToolName(), Version: "test"},
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

	got, err := Text(&report, TextOptions{Explain: true})
	if err != nil {
		t.Fatal(err)
	}
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
	if got := valueColor(
		"not-a-metric",
		1,
		&columnStats{minimum: 0, maximum: 2, count: 2},
	); got != "" {
		t.Fatalf("valueColor = %q, want empty", got)
	}
}

func TestMeanCellNilStats(t *testing.T) {
	cell := meanCell(nil, func(float64) string { return "" })
	if cell.text != naCell {
		t.Fatalf("meanCell(nil) = %q, want %q", cell.text, naCell)
	}
}

// White-box: the guide's direction, boundedness, and column labels mirror
// the renderers' quality map and abbreviations — the guide may never
// contradict how the report actually colors and titles a column.
func TestMetricDocsDirectionMatchesQuality(t *testing.T) {
	t.Parallel()

	for _, d := range MetricDocs() {
		if d.Scope == DocScopeStructural {
			if d.Direction != DirectionNeutral {
				t.Errorf(
					"%s: structural direction = %q, want %q",
					d.Name,
					d.Direction,
					DirectionNeutral,
				)
			}

			continue
		}

		if d.Label != abbrev(d.Name) {
			t.Errorf("%s: label = %q, want column heading %q", d.Name, d.Label, abbrev(d.Name))
		}

		q, colored := qualityForMetric(d.Name)
		if !colored {
			if d.Direction != DirectionNeutral {
				t.Errorf(
					"%s: uncolored metric direction = %q, want %q",
					d.Name,
					d.Direction,
					DirectionNeutral,
				)
			}

			continue
		}

		want := DirectionHigher
		if q.bias == biasLowerBetter {
			want = DirectionLower
		}

		if d.Direction != want {
			t.Errorf("%s: direction = %q, want %q", d.Name, d.Direction, want)
		}

		if d.Bounded != q.bounded {
			t.Errorf("%s: bounded = %v, want %v", d.Name, d.Bounded, q.bounded)
		}
	}
}

func typeMetric(name string, value float64) metrics.MetricResult {
	return metrics.MetricResult{
		Name:       name,
		Scope:      metrics.ScopeType,
		Value:      value,
		Applicable: true,
	}
}

func tableReport() *reusability.Report {
	return &reusability.Report{
		SchemaVersion: "1",
		Tool:          reusability.ToolInfo{Name: reusability.ToolName(), Version: "test"},
		Module:        "example.com/mod",
		Packages: []reusability.PackageReport{{
			Path: "example.com/mod",
			Types: []reusability.TypeReport{
				{Name: "Cart", Reusability: typeMetric(metrics.MetricReusability, 0.80)},
				{Name: "Order", Reusability: metrics.MetricResult{
					Name:       metrics.MetricReusability,
					Scope:      metrics.ScopeType,
					Applicable: false,
					Reason:     "every component dropped",
				}},
			},
		}},
	}
}

func mustMatch(t *testing.T, got, pattern string) {
	t.Helper()

	if !regexp.MustCompile(pattern).MatchString(got) {
		t.Errorf("output does not match %q\ngot:\n%s", pattern, got)
	}
}

func TestTextTreeTableLayout(t *testing.T) {
	got, err := Text(tableReport(), TextOptions{})
	if err != nil {
		t.Fatal(err)
	}

	mustMatch(t, got, `(?m)^module example\.com/mod$`)
	mustMatch(t, got, `(?m)^PATH / TYPE\s+Reuse$`)
	mustMatch(t, got, `(?m)^\.\s+0\.80$`)
	mustMatch(t, got, `(?m)^├── Cart\s+0\.80$`)
	mustMatch(t, got, `(?m)^└── Order\s+–$`)
	mustMatch(t, got, `(?m)^– = not applicable$`)

	if strings.Contains(got, "mean") {
		t.Errorf("output still contains a separate mean row:\n%s", got)
	}

	if strings.Contains(got, "\x1b[") {
		t.Errorf("uncolored output contains ANSI escapes:\n%q", got)
	}
}

func TestTextTreeGroupsPackagesUnderSharedPath(t *testing.T) {
	report := tableReport()
	report.Packages = []reusability.PackageReport{
		{
			Path: "example.com/mod/internal/a",
			Types: []reusability.TypeReport{
				{Name: "T1", Reusability: typeMetric(metrics.MetricReusability, 0.5)},
			},
		},
		{
			Path: "example.com/mod/internal/b/deep",
			Types: []reusability.TypeReport{
				{Name: "T2", Reusability: typeMetric(metrics.MetricReusability, 1)},
			},
		},
	}

	got, err := Text(report, TextOptions{})
	if err != nil {
		t.Fatal(err)
	}

	mustMatch(t, got, `(?m)^internal\s+0\.75$`)
	mustMatch(t, got, `(?m)^├── a\s+0\.50$`)
	mustMatch(t, got, `(?m)^│   └── T1\s+0\.50$`)
	mustMatch(t, got, `(?m)^│$`)
	mustMatch(t, got, `(?m)^└── b/deep\s+1\.00$`)
	mustMatch(t, got, `(?m)^    └── T2\s+1\.00$`)
}

func TestTextReasonsOnlyWithExplain(t *testing.T) {
	report := tableReport()

	got, err := Text(report, TextOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "every component dropped") {
		t.Errorf("reasons shown without Explain:\n%s", got)
	}

	got, err = Text(report, TextOptions{Explain: true})
	if err != nil {
		t.Fatal(err)
	}

	wantLines := []string{
		notesLabel,
		"  example.com/mod",
		"    reusability: every component dropped (Order)",
	}
	for _, line := range wantLines {
		if !strings.Contains(got, line+"\n") {
			t.Errorf("explain output missing line %q\ngot:\n%s", line, got)
		}
	}
}

func TestTextMeanSkipsNonApplicable(t *testing.T) {
	got, err := Text(tableReport(), TextOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(got, "0.40") || strings.Contains(got, "0.4") {
		t.Errorf("mean included a non-applicable value:\n%s", got)
	}
}

func TestTextColorAppliesQualityAndBold(t *testing.T) {
	got, err := Text(tableReport(), TextOptions{Color: true})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, ansiGreen+"0.80"+ansiReset) {
		t.Errorf("high reusability not green:\n%q", got)
	}

	if !strings.Contains(got, ansiBold+ansiGreen+"0.80"+ansiReset) {
		t.Errorf("package-row reusability mean not bold green:\n%q", got)
	}

	if !strings.Contains(got, "├── Cart") {
		t.Errorf("tree glyph missing or styled:\n%q", got)
	}
}

func TestFormatCell(t *testing.T) {
	cases := map[float64]string{
		12:        "12.00",
		0:         "0.00",
		4.25:      "4.25",
		2.0 / 3.0: "0.67",
		0.5:       "0.50",
	}
	for value, want := range cases {
		if got := formatCell(value); got != want {
			t.Errorf("formatCell(%v) = %q, want %q", value, got, want)
		}
	}
}

// Black-box: every reported metric has a complete guide entry, with no
// duplicate names.
func TestMetricDocsCoverEveryMetric(t *testing.T) {
	t.Parallel()

	docs := MetricDocs()

	byName := make(map[string]MetricDoc, len(docs))
	for _, d := range docs {
		if _, dup := byName[d.Name]; dup {
			t.Errorf("duplicate docs entry %q", d.Name)
		}

		byName[d.Name] = d
	}

	for _, name := range metrics.ReportedMetricOrder() {
		assertMetricDoc(t, byName, name, DocScopeType)
	}
}

// assertMetricDoc checks one computed metric's entry for completeness.
func assertMetricDoc(
	t *testing.T,
	byName map[string]MetricDoc,
	name string,
	scope DocScope,
) {
	t.Helper()

	d, ok := byName[name]
	if !ok {
		t.Errorf("metric %q has no docs entry", name)

		return
	}

	if d.Scope != scope {
		t.Errorf("%s scope = %q, want %q", name, d.Scope, scope)
	}

	for field, value := range map[string]string{
		"Label":          d.Label,
		"FullName":       d.FullName,
		"FormulaLaTeX":   d.FormulaLaTeX,
		"Summary":        d.Summary,
		"HowCalculated":  d.HowCalculated,
		"Interpretation": d.Interpretation,
		"Example":        d.Example,
	} {
		if value == "" {
			t.Errorf("%s: %s is empty", name, field)
		}
	}

	if !strings.Contains(d.FormulaMathML, "<math") {
		t.Errorf("%s: FormulaMathML carries no <math> markup", name)
	}

	if !strings.HasPrefix(d.Definition, "reusability/") {
		t.Errorf("%s: Definition = %q, want a versioned reusability id", name, d.Definition)
	}
}

// Black-box: format parsing accepts the known encodings and rejects others.
func TestParseFormat(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"text", "json", "csv", "web"} {
		if f, ok := ParseFormat(name); !ok || string(f) != name {
			t.Errorf("ParseFormat(%q) = %v,%v", name, f, ok)
		}
	}

	if _, ok := ParseFormat("xml"); ok {
		t.Error("xml must be rejected")
	}
}

// Black-box: the text and CSV renderers emit the module and a per-metric
// header/records.
func TestTextAndCSVRendering(t *testing.T) {
	t.Parallel()

	rep := reusability.Report{
		SchemaVersion: "1",
		Tool:          reusability.ToolInfo{Name: reusability.ToolName(), Version: "t"},
		Module:        "example.com/m",
		Packages: []reusability.PackageReport{
			{
				Path: "example.com/m/a",
				Types: []reusability.TypeReport{{Name: "A", Reusability: metrics.MetricResult{
					Name:       metrics.MetricReusability,
					Scope:      metrics.ScopeType,
					Value:      0.7,
					Applicable: true,
				}}},
			},
		},
	}

	text, err := Text(&rep, TextOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "example.com/m") || !strings.Contains(text, "A") {
		t.Errorf("text output missing content:\n%s", text)
	}

	if len(CSVHeader()) == 0 {
		t.Error("empty CSV header")
	}

	if len(CSVRecords(&rep)) == 0 {
		t.Error("no CSV records produced")
	}

	if FormatValue(0.5) == "" {
		t.Error("FormatValue produced empty string")
	}
}
