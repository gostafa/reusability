package reusability

import (
	"testing"

	"github.com/gostafa/reusability/internal/shared/metrics"
)

func TestWithDefaults(t *testing.T) {
	cfg := configWithDefaults(Config{})
	if len(cfg.Patterns) != 1 || cfg.Patterns[0] != "./..." {
		t.Fatalf("patterns = %v", cfg.Patterns)
	}

	if cfg.DependencyScope != DependencyScopeModule {
		t.Fatalf("scope = %q", cfg.DependencyScope)
	}

	if cfg.FieldUsageMode != FieldUsageDirect {
		t.Fatalf("field usage = %q", cfg.FieldUsageMode)
	}

	if cfg.ReusabilityWeights != metrics.DefaultReusabilityWeights() {
		t.Fatalf("weights = %+v", cfg.ReusabilityWeights)
	}
}

func TestValidate(t *testing.T) {
	valid := configWithDefaults(Config{})
	err := validateConfig(valid)
	if err != nil {
		t.Fatal(err)
	}

	bad := valid

	bad.DependencyScope = "galaxy"
	err = validateConfig(bad)
	if err == nil {
		t.Fatal("invalid scope accepted")
	}

	bad = valid

	bad.FieldUsageMode = "psychic"
	err = validateConfig(bad)
	if err == nil {
		t.Fatal("invalid field usage accepted")
	}

	bad = valid

	bad.Patterns = []string{""}
	err = validateConfig(bad)
	if err == nil {
		t.Fatal("empty pattern accepted")
	}

	bad = valid

	bad.ReusabilityWeights = ReusabilityWeights{Cohesion: -0.5, Coupling: 1}
	err = validateConfig(bad)
	if err == nil {
		t.Fatal("negative weight accepted")
	}
}

func TestAllMetrics(t *testing.T) {
	got := AllMetrics()
	if len(got) != 1 || got[0] != MetricReusability {
		t.Fatalf("AllMetrics() = %v, want [%s]", got, MetricReusability)
	}

	if def := DefaultMetrics(); len(def) != 1 || def[0] != MetricReusability {
		t.Fatalf("DefaultMetrics() = %v", def)
	}
}
