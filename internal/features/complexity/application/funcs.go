// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"github.com/gostafa/reusability/internal/features/complexity/domain"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
)

const (
	zero = 0
)

// ComputeForType evaluates cyclomatic complexity for every method of the
// type and derives AMC.
func ComputeForType(t *typefacts.TypeFacts) Result {
	complexities := make([]int, zero, len(t.Methods))
	total := zero

	for index := range t.Methods {
		complexity := domain.Cyclomatic(&t.Methods[index].Branches)

		complexities = append(complexities, complexity)

		total += complexity
	}

	return Result{
		MethodComplexities: complexities,
		AMC:                metrics.AMC(total, len(t.Methods)),
	}
}
