package domain_test

import (
	"math"
	"strings"
	"testing"

	"github.com/gostafa/reusability/internal/features/policy/domain"
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

func naMetric(name string) metrics.MetricResult {
	return metrics.MetricResult{
		Name:       name,
		Scope:      metrics.ScopeType,
		Applicable: false,
		Reason:     "not applicable",
	}
}

func sampleReport() reusability.Report {
	return reusability.Report{
		Packages: []reusability.PackageReport{{
			Path:            "example.com/m/foo",
			Afferent:        2,
			Efferent:        20,
			ExportedFuncs:   20,
			UnexportedFuncs: 20,
			Vars:            20,
			Consts:          20,
			Functions: []reusability.FunctionReport{{
				Name:       "Build",
				Lines:      100,
				Cyclomatic: 12,
			}},
			Types: []reusability.TypeReport{{
				Name:    "Big",
				Fields:  20,
				Methods: 25,
				Metrics: []metrics.MetricResult{
					typeMetric(metrics.MetricReusability, 0.30),
					naMetric("lcom"),
				},
				MethodDetails: []reusability.FunctionReport{{
					Name:       "Do",
					Receiver:   "Big",
					Lines:      90,
					Cyclomatic: 11,
				}},
			}},
		}},
	}
}

func TestEvaluateFlagsMaxAndMinAndSkipsNotApplicable(t *testing.T) {
	t.Parallel()

	policy := domain.Policy{
		TypeMetrics: map[string]domain.Limit{
			metrics.MetricReusability: {Min: 0.6, HasMin: true},
		},
	}
	policy.Package.Efferent = domain.Limit{Max: 15, HasMax: true}
	policy.Package.Funcs.Count = domain.Limit{Max: 35, HasMax: true}
	policy.Package.Vars = domain.Limit{Max: 15, HasMax: true}
	policy.Package.Consts = domain.Limit{Max: 15, HasMax: true}
	policy.Type.Fields = domain.Limit{Max: 12, HasMax: true}
	policy.Funcs.Lines = domain.Limit{Max: 80, HasMax: true}
	policy.Funcs.Cyclomatic = domain.Limit{Max: 10, HasMax: true}

	got := domain.Evaluate(sampleReport(), policy)

	want := []struct {
		typ, fn, key string
		cmp          domain.Comparator
	}{
		{"", "", domain.KeyFuncs, domain.ComparatorMax},
		{"", "", domain.KeyVars, domain.ComparatorMax},
		{"", "", domain.KeyConsts, domain.ComparatorMax},
		{"", "", domain.KeyEfferent, domain.ComparatorMax},
		{"", "Build", domain.KeyFuncLines, domain.ComparatorMax},
		{"", "Build", domain.KeyFuncCyclomatic, domain.ComparatorMax},
		{"Big", "", domain.KeyFields, domain.ComparatorMax},
		{"Big", "Do", domain.KeyFuncLines, domain.ComparatorMax},
		{"Big", "Do", domain.KeyFuncCyclomatic, domain.ComparatorMax},
		{"Big", "", metrics.MetricReusability, domain.ComparatorMin},
	}

	if len(got) != len(want) {
		t.Fatalf("violations = %d, want %d\n%+v", len(got), len(want), got)
	}

	for i, w := range want {
		if got[i].Type != w.typ || got[i].Function != w.fn ||
			got[i].Key != w.key || got[i].Comparator != w.cmp {
			t.Errorf("violation[%d] = (%q %q %q %q), want (%q %q %q %q)",
				i,
				got[i].Type,
				got[i].Function,
				got[i].Key,
				got[i].Comparator,
				w.typ,
				w.fn,
				w.key,
				w.cmp,
			)
		}
	}
}

func TestEvaluateCleanReportHasNoViolations(t *testing.T) {
	t.Parallel()

	clean := reusability.Report{
		Packages: []reusability.PackageReport{{
			Path:          "example.com/m/tidy",
			ExportedFuncs: 3,
			Vars:          1,
			Consts:        1,
			Types: []reusability.TypeReport{{
				Name: "Small", Fields: 2, Methods: 2,
				Metrics: []metrics.MetricResult{typeMetric(metrics.MetricReusability, 0.9)},
			}},
		}},
	}

	if got := domain.Evaluate(clean, domain.DefaultPolicy()); len(got) != 0 {
		t.Errorf("clean report produced violations: %+v", got)
	}
}

func TestEvaluateChecksBothBounds(t *testing.T) {
	t.Parallel()

	report := reusability.Report{
		Packages: []reusability.PackageReport{{
			Path: "p",
			Types: []reusability.TypeReport{{
				Name: "T",
				Metrics: []metrics.MetricResult{
					typeMetric(metrics.MetricReusability, 0.9),
				},
			}},
		}},
	}
	policy := domain.Policy{Metrics: map[string]domain.Limit{
		metrics.MetricReusability: {Min: 0.1, HasMin: true, Max: 0.6, HasMax: true},
	}}

	got := domain.Evaluate(report, policy)
	if len(got) != 1 || got[0].Comparator != domain.ComparatorMax {
		t.Fatalf("want one max violation, got %+v", got)
	}
}

func TestEvaluateToleratesFloatingPointNoiseAtBoundary(t *testing.T) {
	t.Parallel()

	report := reusability.Report{Packages: []reusability.PackageReport{{
		Path: "p",
		Types: []reusability.TypeReport{{
			Name: "AtBoundary",
			Metrics: []metrics.MetricResult{
				typeMetric(metrics.MetricReusability, math.Nextafter(0.5, 0)),
			},
		}},
	}}}
	policy := domain.Policy{TypeMetrics: map[string]domain.Limit{
		metrics.MetricReusability: {Min: 0.5, HasMin: true},
	}}

	if got := domain.Evaluate(report, policy); len(got) != 0 {
		t.Fatalf("adjacent float below boundary produced violations: %+v", got)
	}

	report.Packages[0].Types[0].Metrics[0].Value = 0.5 - 1e-9
	if got := domain.Evaluate(report, policy); len(got) != 1 {
		t.Fatalf("meaningful threshold crossing produced %d violations, want 1", len(got))
	}
}

func TestFormatViolations(t *testing.T) {
	t.Parallel()

	if s := domain.FormatViolations(nil); s != "" {
		t.Errorf("empty slice = %q, want empty", s)
	}

	out := domain.FormatViolations([]domain.Violation{
		{
			Package:    "example.com/m/foo",
			Key:        domain.KeyTypes,
			Value:      25,
			Comparator: domain.ComparatorMax,
			Threshold:  15,
		},
		{
			Package:    "example.com/m/foo",
			Function:   "Build",
			Key:        domain.KeyFuncLines,
			Value:      100,
			Comparator: domain.ComparatorMax,
			Threshold:  80,
		},
		{
			Package:    "example.com/m/foo",
			Type:       "Big",
			Function:   "Do",
			Key:        domain.KeyFuncCyclomatic,
			Value:      12,
			Comparator: domain.ComparatorMax,
			Threshold:  10,
		},
		{
			Package:    "example.com/m/foo",
			Type:       "Big",
			Key:        metrics.MetricReusability,
			Value:      0.42,
			Comparator: domain.ComparatorMin,
			Threshold:  0.6,
		},
	})

	for _, want := range []string{
		"policy: 4 violations",
		"example.com/m/foo (package): types 25 exceeds max 15",
		"example.com/m/foo.Build (func): funcs.lines 100 exceeds max 80",
		"example.com/m/foo.Big.Do (func): funcs.cyclomatic 12 exceeds max 10",
		"example.com/m/foo.Big (type): reusability 0.42 is below min 0.60",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}

	single := domain.FormatViolations(
		[]domain.Violation{
			{
				Package:    "p",
				Key:        domain.KeyExportedFuncs,
				Value:      5,
				Comparator: domain.ComparatorMax,
				Threshold:  3,
			},
		},
	)
	if !strings.HasPrefix(single, "policy: 1 violation\n") {
		t.Errorf("singular header wrong: %q", single)
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	if err := domain.Validate(domain.DefaultPolicy()); err != nil {
		t.Errorf("DefaultPolicy invalid: %v", err)
	}

	cases := map[string]domain.Policy{
		"unknown metric": {Metrics: map[string]domain.Limit{"nope": {Max: 1, HasMax: true}}},
		"min over max": {
			Metrics: map[string]domain.Limit{
				metrics.MetricReusability: {Min: 5, HasMin: true, Max: 2, HasMax: true},
			},
		},
		"hidden metric": {
			TypeMetrics: map[string]domain.Limit{"lcom": {Max: 1, HasMax: true}},
		},
	}
	for name, policy := range cases {
		if err := domain.Validate(policy); err == nil {
			t.Errorf("%s: want error, got nil", name)
		}
	}

	err := domain.Validate(domain.Policy{TypeMetrics: map[string]domain.Limit{
		"lcom": {Max: 0.5, HasMax: true},
	}})
	if err == nil || !strings.Contains(err.Error(), "unknown policy metric") {
		t.Fatalf("hidden metric error = %v, want unknown policy metric", err)
	}
}

func TestApplyOverride(t *testing.T) {
	t.Parallel()

	var policy domain.Policy

	if err := domain.ApplyOverride(&policy, domain.KeyTypes, domain.ComparatorMax, 10); err != nil {
		t.Fatal(err)
	}

	if err := domain.ApplyOverride(
		&policy,
		metrics.MetricReusability,
		domain.ComparatorMin,
		0.7,
	); err != nil {
		t.Fatal(err)
	}

	if err := domain.ApplyOverride(
		&policy,
		"type."+metrics.MetricReusability,
		domain.ComparatorMin,
		0.8,
	); err != nil {
		t.Fatal(err)
	}

	if err := domain.ApplyOverride(&policy, "bogus", domain.ComparatorMax, 1); err == nil {
		t.Error("unknown key: want error, got nil")
	}

	if err := domain.ApplyOverride(&policy, "type.lcom", domain.ComparatorMax, 0.5); err == nil {
		t.Error("hidden metric override: want error, got nil")
	}

	if !policy.Package.Types.HasMax || policy.Package.Types.Max != 10 {
		t.Errorf("types override not applied: %+v", policy.Package.Types)
	}
	if err := domain.ApplyOverride(&policy, domain.KeyFuncs, domain.ComparatorMax, 5); err != nil {
		t.Fatal(err)
	}
	if err := domain.ApplyOverride(&policy, domain.KeyVars, domain.ComparatorMax, 6); err != nil {
		t.Fatal(err)
	}
	if err := domain.ApplyOverride(&policy, domain.KeyConsts, domain.ComparatorMax, 7); err != nil {
		t.Fatal(err)
	}
	if l := policy.Metrics[metrics.MetricReusability]; !l.HasMin || l.Min != 0.7 {
		t.Errorf("reusability override not applied: %+v", l)
	}
	if l := policy.TypeMetrics[metrics.MetricReusability]; !l.HasMin || l.Min != 0.8 {
		t.Errorf("type.reusability override not applied: %+v", l)
	}
}

func TestMetricNamesSorted(t *testing.T) {
	t.Parallel()

	policy := domain.Policy{Metrics: map[string]domain.Limit{
		metrics.MetricReusability: {Min: 0.3, HasMin: true},
	}, TypeMetrics: map[string]domain.Limit{
		metrics.MetricReusability: {Min: 0.7, HasMin: true},
	}}

	got := domain.MetricNames(policy)
	want := []string{metrics.MetricReusability}
	if len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("names = %v, want %v", got, want)
	}
}
