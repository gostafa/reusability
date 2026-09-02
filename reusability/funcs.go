// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package reusability

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/gostafa/reusability/internal/features/projectanalysis/ports/inbound"
	"github.com/gostafa/reusability/internal/infrastructure/analyzer"
	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/internal/shared/version"
)

var (
	errEmptyPackagePattern    = errors.New("empty package pattern")
	errInvalidDependencyScope = errors.New(
		"invalid dependency scope (want project, module, or all)",
	)
	errInvalidFieldUsage = errors.New(
		"invalid field usage mode (want direct or transitive)",
	)
)

// AllMetrics returns every reported metric name. This linter reports one.
func AllMetrics() []MetricName {
	return []MetricName{MetricReusability}
}

// ToolName is embedded in reports as the producing tool.
func ToolName() string {
	return string(MetricReusability)
}

// Analyze validates the configuration, runs the analysis pipeline once over
// the configured patterns, and returns a deterministic report. The context
// cancels package loading and metric computation.
func Analyze(ctx context.Context, config *Config) (Report, error) {
	cfg := configWithDefaults(config)

	err := validateConfig(&cfg)
	if err != nil {
		return Report{}, fmt.Errorf("validate config: %w", err)
	}

	opts := toInbound(&cfg)

	result, analyzeErr := analyzer.NewAnalyzer().Analyze(ctx, &opts)
	if analyzeErr != nil {
		return Report{}, fmt.Errorf("analyze: %w", analyzeErr)
	}

	return toReport(&result), nil
}

// DefaultMetrics returns the reported metric set, which is fixed.
func DefaultMetrics() []MetricName {
	return AllMetrics()
}

func configWithDefaults(cfg *Config) Config {
	out := *cfg

	if len(out.Patterns) == zero {
		out.Patterns = []string{defaultPackagePattern}
	}

	if out.DependencyScope == emptyString {
		out.DependencyScope = DependencyScopeModule
	}

	if out.FieldUsageMode == emptyString {
		out.FieldUsageMode = FieldUsageDirect
	}

	if (out.ReusabilityWeights == Weights{}) {
		out.ReusabilityWeights = metrics.DefaultReusabilityWeights()
	}

	return out
}

func toInbound(cfg *Config) inbound.Options {
	return inbound.Options{
		Directory:            cfg.Directory,
		Patterns:             cfg.Patterns,
		IncludeTests:         cfg.IncludeTests,
		IncludeGenerated:     cfg.IncludeGenerated,
		BuildTags:            cfg.BuildTags,
		Workers:              cfg.Workers,
		DependencyScope:      string(cfg.DependencyScope),
		FieldUsageTransitive: cfg.FieldUsageMode == FieldUsageTransitive,
		ContinueOnError:      cfg.ContinueOnError,
		Weights:              cfg.ReusabilityWeights,
	}
}

func toReport(result *inbound.Result) Report {
	report := Report{
		SchemaVersion: SchemaVersion,
		Tool:          ToolInfo{Name: ToolName(), Version: version.Version()},
		Module:        result.ModulePath,
		Packages:      make([]PackageReport, len(result.Packages)),
	}

	for i := range result.Packages {
		report.Packages[i] = packageReport(&result.Packages[i])
	}

	return report
}

func packageReport(pkg *inbound.PackageResult) PackageReport {
	out := PackageReport{
		Path:  pkg.Path,
		Types: make([]TypeReport, len(pkg.Types)),
	}

	for idx := range pkg.Types {
		named := &pkg.Types[idx]

		out.Types[idx] = TypeReport{Name: named.Name, Reusability: named.Reusability}
	}

	return out
}

func validateConfig(cfg *Config) error {
	err := validateConfigEnums(cfg)
	if err != nil {
		return fmt.Errorf("validate config enums: %w", err)
	}

	err = validateConfigValues(cfg)
	if err != nil {
		return fmt.Errorf("validate config values: %w", err)
	}

	return nil
}

func validateConfigEnums(cfg *Config) error {
	err := validateDependencyScope(cfg.DependencyScope)
	if err != nil {
		return fmt.Errorf("dependency scope: %w", err)
	}

	err = validateFieldUsage(cfg.FieldUsageMode)
	if err != nil {
		return fmt.Errorf("field usage: %w", err)
	}

	return nil
}

func validateConfigValues(cfg *Config) error {
	if slices.Contains(cfg.Patterns, emptyString) {
		return errEmptyPackagePattern
	}

	err := cfg.ReusabilityWeights.Validate()
	if err != nil {
		return fmt.Errorf("reusability weights: %w", err)
	}

	return nil
}

func validateDependencyScope(scope DependencyScope) error {
	switch scope {
	case DependencyScopeProject, DependencyScopeModule, DependencyScopeAll:
		return nil
	default:
		return fmt.Errorf(errWrapQuoted, errInvalidDependencyScope, scope)
	}
}

func validateFieldUsage(mode FieldUsageMode) error {
	switch mode {
	case FieldUsageDirect, FieldUsageTransitive:
		return nil
	default:
		return fmt.Errorf(errWrapQuoted, errInvalidFieldUsage, mode)
	}
}
