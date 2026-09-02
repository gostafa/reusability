package domain

import (
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

	if got := Evaluate(report, DefaultRules()); len(got) != 0 {
		t.Fatalf("Evaluate() = %#v, want no violations", got)
	}
}
