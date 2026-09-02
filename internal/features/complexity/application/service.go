package application

import (
	"github.com/gostafa/reusability/internal/features/complexity/domain"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
)

// Result is the complexity feature's output for one type.
type Result struct {
	// MethodComplexities holds each method's cyclomatic complexity, indexed
	// like TypeFacts.Methods.
	MethodComplexities []int
	// AMC is the type's Average Method Complexity.
	AMC metrics.MetricResult
}

// ComputeForType evaluates cyclomatic complexity for every method of the
// type and derives AMC.
func ComputeForType(t *typefacts.TypeFacts) Result {
	complexities := make([]int, len(t.Methods))
	total := 0

	for i := range t.Methods {
		c := domain.Cyclomatic(t.Methods[i].Branches)
		complexities[i] = c
		total += c
	}

	return Result{
		MethodComplexities: complexities,
		AMC:                metrics.AMC(total, len(t.Methods)),
	}
}
