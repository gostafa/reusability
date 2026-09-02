package metrics

// MetricAMC is the metric name for Average Method Complexity.
const MetricAMC = "amc"

// DefinitionAMC versions the AMC formula implemented by AMC.
const DefinitionAMC = "reusability/amc-v1"

// AMC computes Average Method Complexity for a type:
//
//	AMC = Σ(method complexities) / methodCount
//
// Not applicable when the type has no methods.
func AMC(totalComplexity, methodCount int) MetricResult {
	if methodCount == 0 {
		return notApplicable(MetricAMC, ScopeType, DefinitionAMC, "type has no methods")
	}

	return applicable(MetricAMC, ScopeType, DefinitionAMC,
		float64(totalComplexity)/float64(methodCount))
}
