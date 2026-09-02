package metrics

// MetricLCOM is the metric name for the LCOM cohesion metric.
const MetricLCOM = "lcom"

// DefinitionLCOM versions the LCOM formula implemented by LCOM.
//
// This is the method-field matrix-density variant, chosen because it is
// defined at methodCount == 1 (unlike Henderson-Sellers LCOM*).
const DefinitionLCOM = "reusability/lcom-v1"

// LCOM computes the matrix-density lack-of-cohesion metric:
//
//	LCOM = 1 − totalMethodFieldAccesses / (fieldCount × methodCount)
//
// totalMethodFieldAccesses is the number of 1-cells in the method-field
// matrix (each method contributes each distinct field it uses once). The
// result is in [0, 1]. Not applicable when fieldCount == 0 or
// methodCount == 0.
func LCOM(totalMethodFieldAccesses, fieldCount, methodCount int) MetricResult {
	if fieldCount == 0 {
		return notApplicable(MetricLCOM, ScopeType, DefinitionLCOM, "type has no fields")
	}

	if methodCount == 0 {
		return notApplicable(MetricLCOM, ScopeType, DefinitionLCOM, "type has no methods")
	}

	density := float64(totalMethodFieldAccesses) / (float64(fieldCount) * float64(methodCount))

	return applicable(MetricLCOM, ScopeType, DefinitionLCOM, 1-density)
}
