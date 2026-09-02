// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	cohesion "github.com/gostafa/reusability/internal/features/cohesion/domain"
	complexity "github.com/gostafa/reusability/internal/features/complexity/domain"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
)

const zero = 0

// Compute derives AMC and LCOM from type facts, derives CBO from referenced
// types, assembles the four components (dropping the not-applicable ones),
// and evaluates the index with renormalized weights. fieldUsage is
// "direct" or "transitive".
func Compute(
	tf *typefacts.TypeFacts,
	weights metrics.ReusabilityWeights,
	fieldUsage string,
) Result {
	amc := typeAMC(tf)
	lcom := typeLCOM(tf, fieldUsage)
	cbo := len(tf.ReferencedTypeIDs)

	return Result{
		CBO: metrics.CBO(cbo),
		Reusability: metrics.Reusability(&metrics.ReusabilityInput{
			Cohesion:    metrics.CohesionComponent(&lcom),
			Coupling:    metrics.CouplingComponent(cbo),
			Testability: metrics.TestabilityComponent(&amc),
			Documentation: metrics.DocumentationComponent(
				tf.DocumentedExportedMembers,
				tf.ExportedMembers,
			),
			Weights: weights,
		}),
	}
}

func typeAMC(tf *typefacts.TypeFacts) metrics.MetricResult {
	total := zero

	for index := range tf.Methods {
		total += complexity.Cyclomatic(&tf.Methods[index].Branches)
	}

	return metrics.AMC(total, len(tf.Methods))
}

func typeLCOM(tf *typefacts.TypeFacts, fieldUsage string) metrics.MetricResult {
	sets := cohesion.EffectiveFieldSets(tf, cohesion.FieldUsageMode(fieldUsage))

	return metrics.LCOM(
		cohesion.TotalFieldAccesses(sets),
		len(tf.Fields),
		len(tf.Methods),
	)
}
