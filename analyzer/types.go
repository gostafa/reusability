// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package analyzer

import (
	"context"
	"sync"

	policydomain "github.com/gostafa/reusability/internal/features/policy/domain"
	"github.com/gostafa/reusability/reusability"
)

type (
	reportAnalyzer interface {
		// Analyze evaluates one reusability configuration.
		Analyze(ctx context.Context, cfg *reusability.Config) (reusability.Report, error)
	}

	analyzeFunc func(ctx context.Context, cfg *reusability.Config) (reusability.Report, error)

	runner struct {
		analyzer reportAnalyzer
		err      error
		byPkg    map[string][]policydomain.Violation
		settings Settings
		once     sync.Once
	}

	runResult struct{}

	// RuleSettings is one inline policy rule in analyzer settings.
	RuleSettings struct {
		// Min is the minimum reusability index required for matching types.
		Min *float64 `json:"min"`
		// Pattern is a package-path glob (* one segment, ** any depth).
		Pattern string `json:"pattern"`
	}

	// Settings configures the golangci-lint / go/analysis adapter.
	Settings struct {
		// ReusabilityWeights optionally overrides component score weights.
		ReusabilityWeights *ReusabilityWeightSettings `json:"reusability_weights"`
		// Directory is the module root used for analysis (empty = cwd).
		Directory string `json:"directory"`
		// DependencyScope selects which imports count toward coupling.
		DependencyScope string `json:"dependency_scope"`
		// FieldUsage selects direct vs transitive field-usage tracking.
		FieldUsage string `json:"field_usage"`
		// Patterns are package patterns passed to the loader (default ./...).
		Patterns []string `json:"patterns"`
		// BuildTags are extra build tags for package loading.
		BuildTags []string `json:"build_tags"`
		// Rules are inline policy thresholds; empty uses recommended defaults.
		Rules []RuleSettings `json:"rules"`
		// Workers is the parallel worker count (0 = runtime default).
		Workers int `json:"workers"`
		// Tests includes _test.go packages when true.
		Tests bool `json:"tests"`
		// Generated includes generated files when true.
		Generated bool `json:"generated"`
		// ContinueOnError keeps analyzing after package load failures.
		ContinueOnError bool `json:"continue_on_error"`
	}

	// ReusabilityWeightSettings holds optional per-component weight overrides.
	ReusabilityWeightSettings struct {
		// Cohesion overrides the cohesion component weight.
		Cohesion *float64 `json:"cohesion"`
		// Coupling overrides the coupling component weight.
		Coupling *float64 `json:"coupling"`
		// Testability overrides the testability component weight.
		Testability *float64 `json:"testability"`
		// Documentation overrides the documentation component weight.
		Documentation *float64 `json:"documentation"`
	}
)
