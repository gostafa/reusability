package domain

import (
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
)

// Result is the reusability feature's output for one type.
type Result struct {
	// CBO is the normalized coupling input, reported standalone only when
	// selected.
	CBO metrics.MetricResult
	// Reusability is the experimental reusability index.
	Reusability metrics.MetricResult
}

// Compute derives CBO from the type's referenced-types fact, assembles the
// four components (dropping the not-applicable ones), and evaluates the
// index with renormalized weights.
func Compute(
	t *typefacts.TypeFacts,
	amc, lcom metrics.MetricResult,
	weights metrics.ReusabilityWeights,
) Result {
	cbo := len(t.ReferencedTypeIDs)

	return Result{
		CBO: metrics.CBO(cbo),
		Reusability: metrics.Reusability(
			metrics.CohesionComponent(lcom),
			metrics.CouplingComponent(cbo),
			metrics.TestabilityComponent(amc),
			metrics.DocumentationComponent(t.DocumentedExportedMembers, t.ExportedMembers),
			weights,
		),
	}
}
