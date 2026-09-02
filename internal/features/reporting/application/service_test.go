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
		SchemaVersion: "6",
		Tool:          reusability.ToolInfo{Name: "reusability", Version: "test"},
		Module:        "example.com/m",
		Packages: []reusability.PackageReport{{
			Path: "example.com/m/a",
			Types: []reusability.TypeReport{
				{Name: "A", Reusability: applicable},
				{Name: "B", Reusability: na},
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

	if got["schema_version"] != "6" {
		t.Errorf("schema_version = %v", got["schema_version"])
	}

	pkg := got["packages"].([]any)[0].(map[string]any)
	if _, ok := pkg["afferent"]; ok {
		t.Error("package afferent should not be in JSON schema v6")
	}

	typ := pkg["types"].([]any)[0].(map[string]any)
	reuse := typ["reusability"].(map[string]any)
	if reuse["applicable"] != true {
		t.Error("first reusability entry must be applicable")
	}
	if reuse["value"].(float64) != 0.7 {
		t.Errorf("reusability value = %v", reuse["value"])
	}
	if _, ok := typ["metrics"]; ok {
		t.Error("type metrics map should not appear in JSON schema v6")
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
