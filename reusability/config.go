package reusability

import (
	"errors"
	"fmt"
	"slices"

	"github.com/gostafa/reusability/internal/shared/metrics"
)

// DependencyScope selects which import edges count toward package coupling.
type DependencyScope string

const (
	// DependencyScopeProject counts only imports of other analyzed packages.
	DependencyScopeProject DependencyScope = "project"
	// DependencyScopeModule counts imports of packages in the same module.
	DependencyScopeModule DependencyScope = "module"
	// DependencyScopeAll counts every import, including external modules and
	// the standard library.
	DependencyScopeAll DependencyScope = "all"
)

// FieldUsageMode selects how method→field usage is resolved.
type FieldUsageMode string

const (
	// FieldUsageDirect counts only fields a method body accesses directly.
	FieldUsageDirect FieldUsageMode = "direct"
	// FieldUsageTransitive additionally propagates field usage through calls
	// to sibling methods of the same type, to a fixpoint.
	FieldUsageTransitive FieldUsageMode = "transitive"
)

// MetricName identifies a reported metric.
type MetricName string

// MetricReusability is the only metric this linter reports: a weighted
// composite of cohesion, coupling, testability, and documentation, evaluated
// once per named type. Its inputs (cohesion, average method complexity, and
// coupling between objects) are computed internally and are not reported,
// selectable, or gateable on their own.
const MetricReusability MetricName = metrics.MetricReusability

// ReusabilityWeights aliases the metrics package's weight type so callers can
// configure weights without importing the metrics package directly.
type ReusabilityWeights = metrics.ReusabilityWeights

// AllMetrics returns every reported metric name. This linter reports one.
func AllMetrics() []MetricName {
	return []MetricName{MetricReusability}
}

// DefaultMetrics returns the reported metric set, which is fixed.
func DefaultMetrics() []MetricName {
	return AllMetrics()
}

// Config controls an analysis run. The zero value is usable: defaults are
// applied by Analyze (pattern "./...", module dependency scope, direct field
// usage, and the default reusability weights).
type Config struct {
	// Directory is the working directory for package loading. Empty means the
	// process working directory.
	Directory string
	// Patterns are the package patterns to analyze. Empty means ["./..."].
	Patterns []string
	// IncludeTests also analyzes test files and test packages.
	IncludeTests bool
	// IncludeGenerated also analyzes files carrying the standard
	// "Code generated … DO NOT EDIT." marker.
	IncludeGenerated bool
	// BuildTags are extra build tags for package loading.
	BuildTags []string
	// Workers bounds analysis concurrency. Zero or negative means
	// min(GOMAXPROCS, packageCount).
	Workers int
	// DependencyScope selects the import edges counted by package coupling
	// metrics. Empty means DependencyScopeModule.
	DependencyScope DependencyScope
	// FieldUsageMode selects direct or transitive field-usage resolution.
	// Empty means FieldUsageDirect.
	FieldUsageMode FieldUsageMode
	// ContinueOnError proceeds past packages that fail to load or type-check.
	ContinueOnError bool
	// ReusabilityWeights overrides the reusability component weights. The
	// zero value means the defaults (0.35, 0.25, 0.25, 0.15).
	ReusabilityWeights ReusabilityWeights
}

// configWithDefaults returns a copy of the config with every empty knob
// replaced by its documented default.
func configWithDefaults(c Config) Config {
	if len(c.Patterns) == 0 {
		c.Patterns = []string{"./..."}
	}

	if c.DependencyScope == "" {
		c.DependencyScope = DependencyScopeModule
	}

	if c.FieldUsageMode == "" {
		c.FieldUsageMode = FieldUsageDirect
	}

	if (c.ReusabilityWeights == ReusabilityWeights{}) {
		c.ReusabilityWeights = metrics.DefaultReusabilityWeights()
	}

	return c
}

// validateConfig checks a defaults-applied config.
func validateConfig(c Config) error {
	switch c.DependencyScope {
	case DependencyScopeProject, DependencyScopeModule, DependencyScopeAll:
	default:
		return fmt.Errorf(
			"invalid dependency scope %q (want project, module, or all)",
			c.DependencyScope,
		)
	}

	switch c.FieldUsageMode {
	case FieldUsageDirect, FieldUsageTransitive:
	default:
		return fmt.Errorf(
			"invalid field usage mode %q (want direct or transitive)",
			c.FieldUsageMode,
		)
	}

	if slices.Contains(c.Patterns, "") {
		return errors.New("empty package pattern")
	}

	err := c.ReusabilityWeights.Validate()
	if err != nil {
		return err
	}

	return nil
}
