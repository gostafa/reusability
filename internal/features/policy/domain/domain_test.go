// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"math"
	"strings"
	"testing"

	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/reusability"
)

func TestEvaluateSkipsInapplicableTypeMetric(t *testing.T) {
	report := reusability.Report{Packages: []reusability.PackageReport{{
		Path: "example.com/p",
		Types: []reusability.TypeReport{{
			Name: "T",
			Reusability: metrics.MetricResult{
				Name:       metrics.MetricReusability,
				Scope:      metrics.ScopeType,
				Applicable: false,
			},
		}},
	}}}

	if got := Evaluate(&report, DefaultRules()); len(got) != 0 {
		t.Fatalf("Evaluate() = %#v, want no violations", got)
	}
}

func typeMetric(value float64) metrics.MetricResult {
	return metrics.MetricResult{
		Name:       metrics.MetricReusability,
		Scope:      metrics.ScopeType,
		Value:      value,
		Applicable: true,
	}
}

func naMetric() metrics.MetricResult {
	return metrics.MetricResult{
		Name:       metrics.MetricReusability,
		Scope:      metrics.ScopeType,
		Applicable: false,
		Reason:     "not applicable",
	}
}

func TestMatchPackage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"**", "example.com/m/foo", true},
		{"**", "a", true},
		{"example.com/m/foo", "example.com/m/foo", true},
		{"example.com/m/foo", "example.com/m/bar", false},
		{"**/internal/**", "example.com/m/internal/store", true},
		{"**/internal/**", "example.com/m/store", false},
		{"example.com/*/foo", "example.com/m/foo", true},
		{"example.com/*/foo", "example.com/m/n/foo", false},
		{"**/foo", "example.com/foo", true},
		{"**/foo", "example.com/a/b/foo", true},
	}

	for _, tc := range cases {
		if got := MatchPackage(tc.pattern, tc.path); got != tc.want {
			t.Errorf("MatchPackage(%q, %q) = %v, want %v", tc.pattern, tc.path, got, tc.want)
		}
	}
}

func TestEvaluateMostSpecificRuleWins(t *testing.T) {
	t.Parallel()

	report := reusability.Report{Packages: []reusability.PackageReport{{
		Path: "example.com/m/internal/store",
		Types: []reusability.TypeReport{{
			Name:        "User",
			Reusability: typeMetric(0.55),
		}},
	}}}

	rules := []Rule{
		{Pattern: "**", Min: 0.6},
		{Pattern: "**/internal/**", Min: 0.5},
	}

	got := Evaluate(&report, rules)
	if len(got) != 0 {
		t.Fatalf("violations = %+v, want none", got)
	}
}

func TestMoreSpecific(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		path     string
		rules    []Rule
		wantRule string
		wantMin  float64
	}{
		{
			name:     "literal segments beat wildcard baseline",
			path:     "example.com/m/internal/store",
			rules:    []Rule{{Pattern: "**", Min: 0.6}, {Pattern: "**/internal/**", Min: 0.5}},
			wantRule: "**/internal/**", wantMin: 0.5,
		},
		{
			name: "fewer wildcards break equal literal count",
			path: "example.com/m/store",
			rules: []Rule{
				{Pattern: "example.com/*/store", Min: 0.6},
				{Pattern: "example.com/m/store", Min: 0.5},
			},
			wantRule: "example.com/m/store",
			wantMin:  0.5,
		},
		{
			name:     "fewer wildcards beat longer patterns",
			path:     "example.com/m/internal/store",
			rules:    []Rule{{Pattern: "**/store", Min: 0.6}, {Pattern: "**/**/store", Min: 0.5}},
			wantRule: "**/store", wantMin: 0.6,
		},
		{
			name: "later rule wins exact specificity tie",
			path: "example.com/m/store",
			rules: []Rule{
				{Pattern: "example.com/*/store", Min: 0.6},
				{Pattern: "example.com/*/store", Min: 0.5},
			},
			wantRule: "example.com/*/store",
			wantMin:  0.5,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotMin, gotRule := matchingRule(tc.path, tc.rules)
			if gotRule != tc.wantRule || gotMin != tc.wantMin {
				t.Fatalf(
					"matchingRule() = (%v, %q), want (%v, %q)",
					gotMin,
					gotRule,
					tc.wantMin,
					tc.wantRule,
				)
			}
		})
	}
}

func TestEvaluateSkipsNotApplicable(t *testing.T) {
	t.Parallel()

	report := reusability.Report{Packages: []reusability.PackageReport{{
		Path: "example.com/p",
		Types: []reusability.TypeReport{{
			Name:        "T",
			Reusability: naMetric(),
		}},
	}}}

	if got := Evaluate(&report, DefaultRules()); len(got) != 0 {
		t.Fatalf("n/a type produced violations: %+v", got)
	}
}

func TestEvaluateCleanReportPasses(t *testing.T) {
	t.Parallel()

	report := reusability.Report{Packages: []reusability.PackageReport{{
		Path: "example.com/m/tidy",
		Types: []reusability.TypeReport{{
			Name:        "Small",
			Reusability: typeMetric(0.9),
		}},
	}}}

	if got := Evaluate(&report, DefaultRules()); len(got) != 0 {
		t.Errorf("clean report produced violations: %+v", got)
	}
}

func TestEvaluateToleratesFloatingPointNoiseAtBoundary(t *testing.T) {
	t.Parallel()

	report := reusability.Report{Packages: []reusability.PackageReport{{
		Path: "p",
		Types: []reusability.TypeReport{{
			Name:        "AtBoundary",
			Reusability: typeMetric(math.Nextafter(0.5, 0)),
		}},
	}}}
	rules := []Rule{{Pattern: "**", Min: 0.5}}

	if got := Evaluate(&report, rules); len(got) != 0 {
		t.Fatalf("adjacent float below boundary produced violations: %+v", got)
	}

	report.Packages[0].Types[0].Reusability.Value = 0.5 - 1e-9
	if got := Evaluate(&report, rules); len(got) != 1 {
		t.Fatalf("meaningful threshold crossing produced %d violations, want 1", len(got))
	}
}

func TestFormatViolations(t *testing.T) {
	t.Parallel()

	if s := FormatViolations(nil); s != "" {
		t.Errorf("empty slice = %q, want empty", s)
	}

	out := FormatViolations([]Violation{{
		Package:   "example.com/m/internal/store",
		Type:      "User",
		Value:     0.55,
		Threshold: 0.8,
		Rule:      "**/internal/**",
	}})

	want := "example.com/m/internal/store.User (type): reusability 0.55 is below min 0.80 (rule **/internal/**)"
	if !strings.Contains(out, want) {
		t.Errorf("output missing %q\ngot:\n%s", want, out)
	}

	single := FormatViolations([]Violation{{Package: "p", Type: "T"}})
	if !strings.HasPrefix(single, "policy: 1 violation\n") {
		t.Errorf("singular header wrong: %q", single)
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	if err := Validate(DefaultRules()); err != nil {
		t.Errorf("DefaultRules invalid: %v", err)
	}

	cases := []struct {
		name  string
		rules []Rule
	}{
		{"empty pattern", []Rule{{Pattern: "", Min: 0.5}}},
		{"nan min", []Rule{{Pattern: "**", Min: math.NaN()}}},
		{"min below zero", []Rule{{Pattern: "**", Min: -0.1}}},
		{"min above one", []Rule{{Pattern: "**", Min: 1.1}}},
	}
	for _, tc := range cases {
		if err := Validate(tc.rules); err == nil {
			t.Errorf("%s: want error, got nil", tc.name)
		}
	}
}

func TestDefaultRules(t *testing.T) {
	t.Parallel()

	rules := DefaultRules()
	if len(rules) != 1 || rules[0].Pattern != "**" || rules[0].Min != 0.7 {
		t.Fatalf("DefaultRules() = %+v", rules)
	}
}
