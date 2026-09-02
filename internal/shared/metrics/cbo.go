package metrics

// MetricCBO is the metric name for type-level Coupling Between Objects.
const MetricCBO = "cbo"

// DefinitionCBO versions the CBO formula implemented by CBO.
const DefinitionCBO = "reusability/cbo-v1"

// CBO computes type-level Coupling Between Objects:
//
//	CBO(t) = |ReferencedTypes(t)|
//
// referencedTypeCount is the number of distinct other analyzed named types
// the type references through its fields, method parameters, method returns,
// and embedded types (self-references excluded). Always applicable; the
// value may be 0.
func CBO(referencedTypeCount int) MetricResult {
	return applicable(MetricCBO, ScopeType, DefinitionCBO, float64(referencedTypeCount))
}
