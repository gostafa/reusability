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
		Min     *float64 `json:"min"`
		Pattern string   `json:"pattern"`
	}

	// Settings configures the golangci-lint / go/analysis adapter.
	Settings struct {
		ReusabilityWeights *ReusabilityWeightSettings `json:"reusability_weights"`
		Directory          string                     `json:"directory"`
		DependencyScope    string                     `json:"dependency_scope"`
		FieldUsage         string                     `json:"field_usage"`
		Patterns           []string                   `json:"patterns"`
		BuildTags          []string                   `json:"build_tags"`
		Rules              []RuleSettings             `json:"rules"`
		Workers            int                        `json:"workers"`
		Tests              bool                       `json:"tests"`
		Generated          bool                       `json:"generated"`
		ContinueOnError    bool                       `json:"continue_on_error"`
	}

	// ReusabilityWeightSettings holds optional per-component weight overrides.
	ReusabilityWeightSettings struct {
		Cohesion      *float64 `json:"cohesion"`
		Coupling      *float64 `json:"coupling"`
		Testability   *float64 `json:"testability"`
		Documentation *float64 `json:"documentation"`
	}
)
