// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package metrics

type (
	// Validator checks that a value is internally consistent.
	Validator interface {
		Validate() error
	}

	// Result is a scored metric outcome.
	Result interface {
		Applicable() bool
		Score() float64
	}

	// Component contributes a weighted share of a composite metric.
	Component interface {
		Contribution() float64
	}

	// Weighting supplies reusability component weights.
	Weighting interface {
		Weights() ReusabilityWeights
	}

	// MetricScope classifies whether a metric is type- or package-level.
	MetricScope string
)
