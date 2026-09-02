package metrics

// This file defines the metrics package's abstraction seams: small
// behavioural contracts over the concrete result and weight types. Only
// Validator has an in-package implementer (ReusabilityWeights); the others
// document the shapes callers may depend on instead of the concrete structs.

// Validator is implemented by configuration values that verify their own
// invariants.
type Validator interface {
	Validate() error
}

// Result is the read-only behavioural contract of an evaluated metric:
// whether its value may be read, and the value itself. It mirrors the data
// MetricResult carries as fields.
type Result interface {
	Applicable() bool
	Score() float64
}

// Component is the contract of a single reusability sub-score that
// contributes a weighted value to the composite index.
type Component interface {
	Contribution() float64
}

// Weighting supplies the component weights used to combine reusability
// sub-scores into the final index.
type Weighting interface {
	Weights() ReusabilityWeights
}

var _ Validator = ReusabilityWeights{}
