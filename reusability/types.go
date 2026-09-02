// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package reusability

import (
	"github.com/gostafa/reusability/internal/shared/metrics"
)

type (
	// DependencyScope selects which imports count toward coupling.
	DependencyScope string

	// FieldUsageMode selects direct or transitive field-usage counting.
	FieldUsageMode string

	// MetricName identifies a reported metric.
	MetricName string

	// Weights is the reusability component weight set.
	Weights = metrics.ReusabilityWeights

	// Config configures one Analyze invocation.
	Config struct {
		Directory          string
		DependencyScope    DependencyScope
		FieldUsageMode     FieldUsageMode
		Patterns           []string
		BuildTags          []string
		ReusabilityWeights Weights
		Workers            int
		IncludeTests       bool
		IncludeGenerated   bool
		ContinueOnError    bool
	}
)
