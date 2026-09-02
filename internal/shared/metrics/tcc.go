package metrics

// MetricTCC is the metric name for Tight Class Cohesion.
const MetricTCC = "tcc"

// DefinitionTCC versions the TCC formula implemented by TCC.
const DefinitionTCC = "reusability/tcc-v1"

// TCC computes Tight Class Cohesion. A method pair is connected when the two
// methods share at least one field:
//
//	TCC = connectedPairs / totalPossiblePairs    = k(k−1)/2
//
// Not applicable when methodCount < 2 (totalPossiblePairs == 0).
func TCC(connectedPairs, methodCount int) MetricResult {
	if methodCount < 2 {
		return notApplicable(MetricTCC, ScopeType, DefinitionTCC,
			"type has fewer than 2 methods")
	}

	totalPairs := methodCount * (methodCount - 1) / 2

	return applicable(MetricTCC, ScopeType, DefinitionTCC,
		float64(connectedPairs)/float64(totalPairs))
}
