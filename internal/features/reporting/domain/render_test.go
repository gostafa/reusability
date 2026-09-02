package domain

import (
	"regexp"
	"strings"
	"testing"

	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/reusability"
)

func typeMetric(name string, value float64) metrics.MetricResult {
	return metrics.MetricResult{
		Name:       name,
		Scope:      metrics.ScopeType,
		Value:      value,
		Applicable: true,
	}
}

func tableReport() reusability.Report {
	return reusability.Report{
		SchemaVersion: "1",
		Tool:          reusability.ToolInfo{Name: "reusability", Version: "test"},
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
	got := Text(tableReport(), TextOptions{})

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

	got := Text(report, TextOptions{})

	mustMatch(t, got, `(?m)^internal\s+0\.75$`)
	mustMatch(t, got, `(?m)^├── a\s+0\.50$`)
	mustMatch(t, got, `(?m)^│   └── T1\s+0\.50$`)
	mustMatch(t, got, `(?m)^│$`)
	mustMatch(t, got, `(?m)^└── b/deep\s+1\.00$`)
	mustMatch(t, got, `(?m)^    └── T2\s+1\.00$`)
}

func TestTextReasonsOnlyWithExplain(t *testing.T) {
	report := tableReport()

	if got := Text(report, TextOptions{}); strings.Contains(got, "every component dropped") {
		t.Errorf("reasons shown without Explain:\n%s", got)
	}

	got := Text(report, TextOptions{Explain: true})

	wantLines := []string{
		"notes",
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
	got := Text(tableReport(), TextOptions{})

	if strings.Contains(got, "0.40") || strings.Contains(got, "0.4") {
		t.Errorf("mean included a non-applicable value:\n%s", got)
	}
}

func TestTextColorAppliesQualityAndBold(t *testing.T) {
	got := Text(tableReport(), TextOptions{Color: true})

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
