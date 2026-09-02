package plugin_test

import (
	"testing"

	"github.com/golangci/plugin-module-register/register"
	"github.com/gostafa/reusability/analyzer"
	"github.com/gostafa/reusability/plugin"
)

func TestNewBuildAnalyzersAndLoadMode(t *testing.T) {
	t.Parallel()

	p, err := plugin.New(map[string]any{
		"dependency-scope": "module",
		"field-usage":      "direct",
		"patterns":         []any{"./..."},
		"reusability-weights": map[string]any{
			"cohesion":      0.1,
			"coupling":      0.2,
			"testability":   0.3,
			"documentation": 0.4,
		},
		"rules": []any{
			map[string]any{"pattern": "**/internal/**", "min": 0.8},
			map[string]any{"pattern": "**", "min": 0.6},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if mode := p.GetLoadMode(); mode != register.LoadModeTypesInfo {
		t.Fatalf("GetLoadMode = %q, want %q", mode, register.LoadModeTypesInfo)
	}

	analyzers, err := p.BuildAnalyzers()
	if err != nil {
		t.Fatal(err)
	}

	if len(analyzers) != 1 {
		t.Fatalf("len(analyzers) = %d, want 1", len(analyzers))
	}

	if analyzers[0].Name != analyzer.Name {
		t.Fatalf("Name = %q, want %q", analyzers[0].Name, analyzer.Name)
	}
}

func TestNewNilSettings(t *testing.T) {
	t.Parallel()

	p, err := plugin.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := p.BuildAnalyzers(); err != nil {
		t.Fatal(err)
	}
}

func TestNewAcceptsPartialReusabilityWeights(t *testing.T) {
	t.Parallel()

	p, err := plugin.New(map[string]any{
		"reusability-weights": map[string]any{
			"coupling": 0,
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

	_, err := plugin.New(map[string]any{"not-a-real-setting": true})
	if err == nil {
		t.Fatal("expected error for unknown settings key")
	}
}

func TestNewRejectsPolicyFileSetting(t *testing.T) {
	t.Parallel()

	_, err := plugin.New(map[string]any{"config": ".modularity.yml"})
	if err == nil {
		t.Fatal("expected error for removed config file setting")
	}
}

func TestNewRejectsUnknownInlinePolicySettings(t *testing.T) {
	t.Parallel()

	cases := map[string]any{
		"unknown rule key": map[string]any{
			"rules": []any{map[string]any{"pattern": "**", "maximum": 1}},
		},
		"reusability weight key": map[string]any{
			"reusability-weights": map[string]any{"collaboration": 1},
		},
	}

	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := plugin.New(raw); err == nil {
				t.Fatal("expected decoding error")
			}
		})
	}
}

func TestNewRejectsInvalidAnalyzerSettings(t *testing.T) {
	t.Parallel()

	p, err := plugin.New(map[string]any{"dependency-scope": "bogus"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.BuildAnalyzers()
	if err == nil {
		t.Fatal("expected BuildAnalyzers error for invalid dependency-scope")
	}

	p, err = plugin.New(map[string]any{
		"rules": []any{map[string]any{"pattern": "**", "min": 2}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err = p.BuildAnalyzers(); err == nil {
		t.Fatal("expected BuildAnalyzers error for invalid rule min")
	}

	p, err = plugin.New(map[string]any{
		"reusability-weights": map[string]any{
			"cohesion":      -1,
			"coupling":      0,
			"testability":   0,
			"documentation": 0,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err = p.BuildAnalyzers(); err == nil {
		t.Fatal("expected BuildAnalyzers error for invalid reusability weights")
	}
}
