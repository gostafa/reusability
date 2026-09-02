// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package metrics

type (
	// MetricResult is one computed metric value with applicability metadata.
	MetricResult struct {
		Name       string
		Scope      MetricScope
		Reason     string
		Definition string
		Value      float64
		Applicable bool
	}

	// ReusabilityWeights holds the four component weights for the index.
	ReusabilityWeights struct {
		Cohesion      float64 // weight of the cohesion component (from LCOM)
		Coupling      float64 // weight of the coupling component (from CBO)
		Testability   float64 // weight of the testability component (from AMC)
		Documentation float64 // weight of the documentation component
	}

	// ReusabilityComponent is one weighted input to the reusability index.
	ReusabilityComponent struct {
		Name       string
		Reason     string
		Value      float64
		Applicable bool
	}

	// ReusabilityInput bundles the four components and their weights.
	ReusabilityInput struct {
		Cohesion      ReusabilityComponent
		Coupling      ReusabilityComponent
		Testability   ReusabilityComponent
		Documentation ReusabilityComponent
		Weights       ReusabilityWeights
	}
)
