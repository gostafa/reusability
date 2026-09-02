package metrics

// MetricScope identifies the kind of entity a metric describes.
type MetricScope string

const (
	// ScopeType marks a metric computed per named type.
	ScopeType MetricScope = "type"
	// ScopePackage marks a metric computed once per package.
	ScopePackage MetricScope = "package"
)

// MetricResult is the outcome of evaluating one metric for one entity.
//
// When a metric is undefined for an entity (for example AMC on a type with no
// methods), Applicable is false and Reason explains why — the Value must not
// be read. A misleading zero is never returned.
type MetricResult struct {
	Name       string      // metric name, e.g. "amc"
	Scope      MetricScope // kind of entity the metric describes
	Value      float64     // metric value; meaningless unless Applicable
	Applicable bool        // whether Value may be read
	Reason     string      // why not applicable, or which components were dropped
	Definition string      // versioned definition, e.g. "modularity/amc-v1"
}

func notApplicable(name string, scope MetricScope, definition, reason string) MetricResult {
	return MetricResult{
		Name:       name,
		Scope:      scope,
		Applicable: false,
		Reason:     reason,
		Definition: definition,
	}
}

func applicable(name string, scope MetricScope, definition string, value float64) MetricResult {
	return MetricResult{
		Name:       name,
		Scope:      scope,
		Value:      value,
		Applicable: true,
		Definition: definition,
	}
}
