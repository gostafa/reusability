// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"github.com/gostafa/reusability/internal/features/cohesion/domain"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
)

// ComputeForType evaluates all cohesion metrics for one type.
func ComputeForType(t *typefacts.TypeFacts, mode domain.FieldUsageMode) Result {
	methodCount := len(t.Methods)
	fieldCount := len(t.Fields)
	sets := domain.EffectiveFieldSets(t, mode)
	sharingPairs := domain.CountSharingPairs(sets, fieldCount)
	accesses := domain.TotalFieldAccesses(sets)

	return Result{
		LCOM: metrics.LCOM(accesses, fieldCount, methodCount),
		TCC:  metrics.TCC(sharingPairs, methodCount),
	}
}
