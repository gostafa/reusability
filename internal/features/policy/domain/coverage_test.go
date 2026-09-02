package domain

import (
	"math"
	"testing"

	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/reusability"
)

func TestEvaluateSkipsInapplicableTypeMetric(t *testing.T) {
	report := reusability.Report{Packages: []reusability.PackageReport{{
		Path: "example.com/p",
		Types: []reusability.TypeReport{{
			Name: "T",
			Metrics: []metrics.MetricResult{{
				Name:       metrics.MetricReusability,
				Scope:      metrics.ScopeType,
				Applicable: false,
			}},
		}},
	}}}
	policy := Policy{TypeMetrics: map[string]Limit{
		metrics.MetricReusability: {Min: 0.7, HasMin: true},
	}}

	if got := Evaluate(report, policy); len(got) != 0 {
		t.Fatalf("Evaluate() = %#v, want no violations", got)
	}
}

func TestValidateRejectsNonFiniteStructuralBounds(t *testing.T) {
	for _, tc := range []struct {
		name   string
		policy Policy
	}{
		{"nan max", Policy{Package: PackageLimits{Types: Limit{Max: math.NaN(), HasMax: true}}}},
		{"inf min", Policy{Type: TypeLimits{Fields: Limit{Min: math.Inf(1), HasMin: true}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := Validate(tc.policy); err == nil {
				t.Fatal("Validate succeeded, want non-finite error")
			}
		})
	}
}

func TestApplyOverrideStructuralKeysAndMin(t *testing.T) {
	var policy Policy
	for _, key := range []string{
		KeyTypes, KeyFuncs, KeyExportedFuncs, KeyUnexportedFuncs, KeyVars, KeyConsts, KeyAfferent, KeyEfferent, KeyFields, KeyMethods,
	} {
		if err := ApplyOverride(&policy, key, ComparatorMin, 1); err != nil {
			t.Fatalf("ApplyOverride(%q, min) = %v", key, err)
		}
	}

	if !policy.Package.Types.HasMin || policy.Package.Types.Min != 1 {
		t.Fatalf("types min = %+v", policy.Package.Types)
	}
	if !policy.Package.Funcs.Count.HasMin || policy.Package.Funcs.Count.Min != 1 {
		t.Fatalf("funcs min = %+v", policy.Package.Funcs)
	}
	if !policy.Type.Methods.HasMin || policy.Type.Methods.Min != 1 {
		t.Fatalf("methods min = %+v", policy.Type.Methods)
	}
}

func TestApplyOverrideUnknownScopedMetrics(t *testing.T) {
	var policy Policy
	if err := ApplyOverride(&policy, "package.nope", ComparatorMax, 1); err == nil {
		t.Fatal("unknown package metric succeeded")
	}
	if err := ApplyOverride(&policy, "type.nope", ComparatorMax, 1); err == nil {
		t.Fatal("unknown type metric succeeded")
	}
}
