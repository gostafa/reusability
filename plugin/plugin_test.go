// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package plugin

import (
	"testing"

	"github.com/golangci/plugin-module-register/register"
	"github.com/gostafa/reusability/analyzer"
)

const (
	extTestZero            = 0
	extTestOne             = 1
	extWeightCohesion      = 0.1
	extWeightCoupling      = 0.2
	extWeightTestability   = 0.3
	extWeightDocumentation = 0.4
	extInternalWildcard    = "**/internal/**"
	extWildcardPattern     = "**"
	extInternalRuleMin     = 0.8
	extDefaultRuleMin      = 0.6
	extInvalidRuleMin      = 2.0
	extNegativeCohesion    = -1
)
const rulePatternKey = "pattern"

func defaultPluginSettings() map[string]any {
	return map[string]any{
		"dependency-scope": "module",
		"field-usage":      "direct",
		"patterns":         []any{"./..."},
		"reusability-weights": map[string]any{
			"cohesion":      extWeightCohesion,
			"coupling":      extWeightCoupling,
			"testability":   extWeightTestability,
			"documentation": extWeightDocumentation,
		},
		"rules": []any{
			map[string]any{rulePatternKey: extInternalWildcard, "min": extInternalRuleMin},
			map[string]any{rulePatternKey: extWildcardPattern, "min": extDefaultRuleMin},
		},
	}
}

func assertBuiltAnalyzer(t *testing.T, p register.LinterPlugin) {
	t.Helper()

	if mode := p.GetLoadMode(); mode != register.LoadModeTypesInfo {
		t.Fatalf("GetLoadMode = %q, want %q", mode, register.LoadModeTypesInfo)
	}

	analyzers, err := p.BuildAnalyzers()
	if err != nil {
		t.Fatal(err)
	}

	if len(analyzers) != extTestOne {
		t.Fatalf("len(analyzers) = %d, want 1", len(analyzers))
	}

	if analyzers[extTestZero].Name != analyzer.Name {
		t.Fatalf("Name = %q, want %q", analyzers[extTestZero].Name, analyzer.Name)
	}
}

func TestNewBuildAnalyzersAndLoadMode(t *testing.T) {
	t.Parallel()

	p, err := New(defaultPluginSettings())
	if err != nil {
		t.Fatal(err)
	}

	assertBuiltAnalyzer(t, p)
}

func TestNewNilSettings(t *testing.T) {
	t.Parallel()

	p, err := New(nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := p.BuildAnalyzers(); err != nil {
		t.Fatal(err)
	}
}

func TestNewAcceptsPartialReusabilityWeights(t *testing.T) {
	t.Parallel()

	p, err := New(map[string]any{
		"reusability-weights": map[string]any{
			"coupling": extTestZero,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := p.BuildAnalyzers(); err != nil {
		t.Fatal(err)
	}
}

func TestNewRejectsUnknownSettings(t *testing.T) {
	t.Parallel()

	_, err := New(map[string]any{"not-a-real-setting": true})
	if err == nil {
		t.Fatal("expected error for unknown settings key")
	}
}

func TestNewRejectsPolicyFileSetting(t *testing.T) {
	t.Parallel()

	_, err := New(map[string]any{"config": ".modularity.yml"})
	if err == nil {
		t.Fatal("expected error for removed config file setting")
	}
}

func TestNewRejectsUnknownInlinePolicySettings(t *testing.T) {
	t.Parallel()

	cases := map[string]any{
		"unknown rule key": map[string]any{
			"rules": []any{
				map[string]any{rulePatternKey: extWildcardPattern, "maximum": extTestOne},
			},
		},
		"reusability weight key": map[string]any{
			"reusability-weights": map[string]any{"collaboration": extTestOne},
		},
	}

	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := New(raw); err == nil {
				t.Fatal("expected decoding error")
			}
		})
	}
}

func assertBuildAnalyzersError(t *testing.T, settings map[string]any) {
	t.Helper()

	p, err := New(settings)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = p.BuildAnalyzers(); err == nil {
		t.Fatal("expected BuildAnalyzers error")
	}
}

func TestNewRejectsInvalidAnalyzerSettings(t *testing.T) {
	t.Parallel()

	assertBuildAnalyzersError(t, map[string]any{"dependency-scope": "bogus"})

	assertBuildAnalyzersError(t, map[string]any{
		"rules": []any{
			map[string]any{rulePatternKey: extWildcardPattern, "min": extInvalidRuleMin},
		},
	})

	assertBuildAnalyzersError(t, map[string]any{
		"reusability-weights": map[string]any{
			"cohesion":      extNegativeCohesion,
			"coupling":      extTestZero,
			"testability":   extTestZero,
			"documentation": extTestZero,
		},
	})
}
