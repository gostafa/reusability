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

type (
	// MetricResult is one computed metric value with applicability metadata.
	MetricResult = struct {
		// Name is the metric identifier (for example "reusability").
		Name string
		// Scope is whether the metric is type- or package-level.
		Scope MetricScope
		// Reason explains drops or non-applicability when set.
		Reason string
		// Definition identifies the formula version used.
		Definition string
		// Value is the numeric score when Applicable is true.
		Value float64
		// Applicable is false when the metric cannot be computed.
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
	ReusabilityComponent = struct {
		// Name identifies the component (cohesion, coupling, …).
		Name string
		// Reason explains why the component was dropped when not applicable.
		Reason string
		// Value is the component score in [0, 1] when Applicable is true.
		Value float64
		// Applicable is false when the component must be dropped.
		Applicable bool
	}

	// ReusabilityInput bundles the four components and their weights.
	ReusabilityInput = struct {
		// Cohesion is the cohesion component derived from LCOM.
		Cohesion ReusabilityComponent
		// Coupling is the coupling component derived from CBO.
		Coupling ReusabilityComponent
		// Testability is the testability component derived from AMC.
		Testability ReusabilityComponent
		// Documentation is the documentation-coverage component.
		Documentation ReusabilityComponent
		// Weights are the four component weights before renormalization.
		Weights ReusabilityWeights
	}
)

type (
	reusabilityResultInput = struct {
		spec      metricSpec
		inputs    []weightedComponent
		dropped   []ReusabilityComponent
		details   []string
		names     []string
		weightSum float64
	}

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
	MetricScope = string
)
