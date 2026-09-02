// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"github.com/gostafa/reusability/internal/shared/metrics"
)

// Compute derives CBO from the type's referenced-types fact, assembles the
// four components (dropping the not-applicable ones), and evaluates the
// index with renormalized weights.
func Compute(input *ComputeInput) Result {
	cbo := len(input.Type.ReferencedTypeIDs)

	return Result{
		CBO: metrics.CBO(cbo),
		Reusability: metrics.Reusability(&metrics.ReusabilityInput{
			Cohesion:    metrics.CohesionComponent(&input.LCOM),
			Coupling:    metrics.CouplingComponent(cbo),
			Testability: metrics.TestabilityComponent(&input.AMC),
			Documentation: metrics.DocumentationComponent(
				input.Type.DocumentedExportedMembers,
				input.Type.ExportedMembers,
			),
			Weights: input.Weights,
		}),
	}
}
