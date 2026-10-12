// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

// Package metrics defines reusable metric formulas and result types.
package metrics

type (
	metricSpec = struct {
		name       string
		scope      MetricScope
		definition string
	}

	weightedComponent = struct {
		component ReusabilityComponent
		weight    float64
	}
)
