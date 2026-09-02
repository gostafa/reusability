package analyzer

import (
	"cmp"
	"fmt"

	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/reusability"
)

// Settings configures the reusability policy analyzer. Analysis fields map to
// the reusability.Config facade. Policy rules are decoded directly from
// golangci-lint's linters.settings.custom.reusability.settings block.
type Settings struct {
	// Directory is the working directory for package loading. Empty means the
	// process working directory.
	Directory string `json:"directory"`
	// Patterns are the package patterns to analyze. Empty means ["./..."].
	Patterns []string `json:"patterns"`
	// Tests includes test files and test packages.
	Tests bool `json:"tests"`
	// Generated includes files with the standard generated-code marker.
	Generated bool `json:"generated"`
	// DependencyScope is "project", "module", or "all". Empty means "module".
	DependencyScope string `json:"dependency-scope"`
	// FieldUsage is "direct" or "transitive". Empty means "direct".
	FieldUsage string `json:"field-usage"`
	// Workers bounds analysis concurrency. Zero selects the facade default.
	Workers int `json:"workers"`
	// ContinueOnError skips packages that fail to load or type-check.
	ContinueOnError bool `json:"continue-on-error"`
	// BuildTags are extra build tags for package loading.
	BuildTags []string `json:"build-tags"`
	// ReusabilityWeights overrides the reusability component weights. Omitted
	// fields keep their defaults; explicit zero values are allowed.
	ReusabilityWeights *ReusabilityWeightSettings `json:"reusability-weights"`
	// Rules maps package-path glob patterns to minimum type-level reusability
	// thresholds. Empty selects DefaultRules() (catch-all min 0.7).
	Rules []RuleSettings `json:"rules"`
}

// ReusabilityWeightSettings configures the component weights used by the
// reusability metric. Pointer fields distinguish omitted settings from an
// explicit zero.
type ReusabilityWeightSettings struct {
	Cohesion      *float64 `json:"cohesion"`
	Coupling      *float64 `json:"coupling"`
	Testability   *float64 `json:"testability"`
	Documentation *float64 `json:"documentation"`
}

func (s Settings) withDefaults() Settings {
	if len(s.Patterns) == 0 {
		s.Patterns = []string{"./..."}
	}

	s.DependencyScope = cmp.Or(s.DependencyScope, string(reusability.DependencyScopeModule))
	s.FieldUsage = cmp.Or(s.FieldUsage, string(reusability.FieldUsageDirect))

	return s
}

func (s Settings) validate() error {
	if err := validateDependencyScope(s.DependencyScope); err != nil {
		return err
	}

	if err := validateFieldUsage(s.FieldUsage); err != nil {
		return err
	}

	if err := s.reusabilityWeights().Validate(); err != nil {
		return err
	}

	_, err := s.rules()

	return err
}

func validateDependencyScope(value string) error {
	switch reusability.DependencyScope(value) {
	case reusability.DependencyScopeProject,
		reusability.DependencyScopeModule,
		reusability.DependencyScopeAll:
		return nil
	default:
		return fmt.Errorf(
			"invalid dependency-scope %q (want project, module, or all)",
			value,
		)
	}
}

func validateFieldUsage(value string) error {
	switch reusability.FieldUsageMode(value) {
	case reusability.FieldUsageDirect, reusability.FieldUsageTransitive:
		return nil
	default:
		return fmt.Errorf("invalid field-usage %q (want direct or transitive)", value)
	}
}

func (s Settings) toConfig() reusability.Config {
	return reusability.Config{
		Directory:          s.Directory,
		Patterns:           append([]string(nil), s.Patterns...),
		IncludeTests:       s.Tests,
		IncludeGenerated:   s.Generated,
		BuildTags:          append([]string(nil), s.BuildTags...),
		Workers:            s.Workers,
		DependencyScope:    reusability.DependencyScope(s.DependencyScope),
		FieldUsageMode:     reusability.FieldUsageMode(s.FieldUsage),
		ContinueOnError:    s.ContinueOnError,
		ReusabilityWeights: s.reusabilityWeights(),
	}
}

func (s Settings) reusabilityWeights() metrics.ReusabilityWeights {
	weights := metrics.DefaultReusabilityWeights()
	if s.ReusabilityWeights == nil {
		return weights
	}

	if s.ReusabilityWeights.Cohesion != nil {
		weights.Cohesion = *s.ReusabilityWeights.Cohesion
	}
	if s.ReusabilityWeights.Coupling != nil {
		weights.Coupling = *s.ReusabilityWeights.Coupling
	}
	if s.ReusabilityWeights.Testability != nil {
		weights.Testability = *s.ReusabilityWeights.Testability
	}
	if s.ReusabilityWeights.Documentation != nil {
		weights.Documentation = *s.ReusabilityWeights.Documentation
	}

	return weights
}
