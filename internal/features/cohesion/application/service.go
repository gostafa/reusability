package application

import (
	"github.com/gostafa/reusability/internal/features/cohesion/domain"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
)

// Result is the cohesion feature's output for one type.
type Result struct {
	// LCOM is the method-field matrix sparsity.
	LCOM metrics.MetricResult
	// TCC is the connected method-pair density.
	TCC metrics.MetricResult
}

// ComputeForType evaluates all cohesion metrics for one type. transitive
// selects the transitive field-usage mode.
func ComputeForType(t *typefacts.TypeFacts, transitive bool) Result {
	methodCount := len(t.Methods)
	fieldCount := len(t.Fields)

	sets := domain.EffectiveFieldSets(t, transitive)
	sharingPairs := domain.CountSharingPairs(sets, fieldCount)
	accesses := domain.TotalFieldAccesses(sets)

	return Result{
		LCOM: metrics.LCOM(accesses, fieldCount, methodCount),
		TCC:  metrics.TCC(sharingPairs, methodCount),
	}
}
