package domain_test

import (
	"strings"
	"testing"

	"github.com/gostafa/reusability/internal/features/reporting/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
)

// Black-box: every reported metric and structural column has a complete
// guide entry, with no duplicate names.
func TestMetricDocsCoverEveryMetric(t *testing.T) {
	t.Parallel()

	docs := domain.MetricDocs()

	byName := make(map[string]domain.MetricDoc, len(docs))
	for _, d := range docs {
		if _, dup := byName[d.Name]; dup {
			t.Errorf("duplicate docs entry %q", d.Name)
		}

		byName[d.Name] = d
	}

	for _, name := range metrics.ReportedMetricOrder() {
		assertMetricDoc(t, byName, name, domain.DocScopeType)
	}

	for _, name := range []string{"ca", "ce", "funcs", "vars", "consts", "types", "fields", "methods"} {
		d, ok := byName[name]
		if !ok {
			t.Errorf("structural column %q has no docs entry", name)

			continue
		}

		if d.Scope != domain.DocScopeStructural {
			t.Errorf("%s scope = %q, want %q", name, d.Scope, domain.DocScopeStructural)
		}

		if d.Label == "" || d.FullName == "" || d.Summary == "" ||
			d.HowCalculated == "" || d.Interpretation == "" || d.Example == "" {
			t.Errorf("structural entry %q has empty prose fields", name)
		}

		// Structural columns are counted, not computed: no formula, no id.
		if d.FormulaMathML != "" || d.Definition != "" {
			t.Errorf("structural entry %q must not carry a formula or definition", name)
		}
	}
}

// assertMetricDoc checks one computed metric's entry for completeness.
func assertMetricDoc(
	t *testing.T,
	byName map[string]domain.MetricDoc,
	name string,
	scope domain.DocScope,
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
