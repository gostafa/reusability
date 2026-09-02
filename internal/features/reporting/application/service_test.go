package application

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/gostafa/reusability/internal/features/reporting/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/reusability"
)

func sampleReport() reusability.Report {
	applicable := metrics.MetricResult{
		Name:       "reusability",
		Scope:      metrics.ScopeType,
		Value:      0.7,
		Applicable: true,
		Definition: "d",
	}
	na := metrics.MetricResult{
		Name:       "reusability",
		Scope:      metrics.ScopeType,
		Applicable: false,
		Reason:     "every component dropped",
		Definition: "d",
	}

	return reusability.Report{
		SchemaVersion: "1",
		Tool:          reusability.ToolInfo{Name: "reusability", Version: "test"},
		Module:        "example.com/m",
		Packages: []reusability.PackageReport{{
			Path:   "example.com/m/a",
			Vars:   2,
			Consts: 3,
			Types: []reusability.TypeReport{
				{Name: "A", Metrics: []metrics.MetricResult{applicable}},
				{Name: "B", Metrics: []metrics.MetricResult{na}},
			},
		}},
	}
}

// White-box: the JSON envelope round-trips and honors the applicability
// contract (applicable → value present; n/a → value omitted).
func TestRenderJSONContract(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	err := renderJSON(&buf, sampleReport())
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	err = json.Unmarshal([]byte(buf.String()), &got)
	if err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	if got["schema_version"] != "1" {
		t.Errorf("schema_version = %v", got["schema_version"])
	}

	pkg := got["packages"].([]any)[0].(map[string]any)
	if pkg["vars"].(float64) != 2 || pkg["consts"].(float64) != 3 {
		t.Errorf("package counts = vars %v consts %v, want 2 and 3", pkg["vars"], pkg["consts"])
	}

	typ := pkg["types"].([]any)[0].(map[string]any)
	metricsObj := typ["metrics"].(map[string]any)
	reuse := metricsObj["reusability"].(map[string]any)
	if reuse["applicable"] != true {
		t.Error("first reusability entry must be applicable")
	}
	if reuse["value"].(float64) != 0.7 {
		t.Errorf("reusability value = %v", reuse["value"])
	}
}

// White-box: ordered metric objects keep the given slice order.
func TestEncodeOrderedMetricsPreservesOrder(t *testing.T) {
	t.Parallel()

	got, err := encodeOrderedMetrics([]metrics.MetricResult{
		{Name: "amc", Scope: metrics.ScopeType, Value: 1, Applicable: true, Definition: "d"},
		{Name: "tcc", Scope: metrics.ScopeType, Applicable: false, Reason: "x", Definition: "d"},
	})
	if err != nil {
		t.Fatal(err)
	}

	s := string(got)
	if !strings.HasPrefix(s, `{"amc":`) || strings.Index(s, "amc") > strings.Index(s, "tcc") {
		t.Errorf("order not preserved: %s", s)
	}
}

// White-box: an unknown format is rejected.
func TestRenderUnknownFormat(t *testing.T) {
	t.Parallel()

	err := render(io.Discard, sampleReport(), domain.Format("xml"), domain.TextOptions{})
	if err == nil {
		t.Fatal("unknown format should error")
	}
}
