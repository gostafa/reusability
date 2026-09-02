// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	cohesion "github.com/gostafa/reusability/internal/features/cohesion/domain"
	complexity "github.com/gostafa/reusability/internal/features/complexity/domain"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
)

const (
	zero = 0
)

// Compute derives AMC and LCOM from type facts, derives CBO from referenced
// types, assembles the four components (dropping the not-applicable ones),
// and evaluates the index with renormalized weights. fieldUsage is
// "direct" or "transitive".
func Compute(
	typeFacts *typefacts.TypeFacts,
	weights *metrics.ReusabilityWeights,
	fieldUsage string,
) Result {
	amc := typeAMC(typeFacts)
	lcom := typeLCOM(typeFacts, fieldUsage)
	cbo := len(typeFacts.ReferencedTypeIDs)

	return Result{
		CBO: metrics.CBO(cbo),
		Reusability: metrics.Reusability(&metrics.ReusabilityInput{
			Cohesion:      metrics.CohesionComponent(&lcom),
			Coupling:      metrics.CouplingComponent(cbo),
			Testability:   metrics.TestabilityComponent(&amc),
			Documentation: documentationComponent(typeFacts),
			Weights:       *weights,
		}),
	}
}

func documentationComponent(typeFacts *typefacts.TypeFacts) metrics.ReusabilityComponent {
	return metrics.DocumentationComponent(
		typeFacts.DocumentedExportedMembers,
		typeFacts.ExportedMembers,
	)
}

func typeAMC(typeFacts *typefacts.TypeFacts) metrics.MetricResult {
	total := zero

	for index := range typeFacts.Methods {
		total += complexity.Cyclomatic(&typeFacts.Methods[index].Branches)
	}

	return metrics.AMC(total, len(typeFacts.Methods))
}

func typeLCOM(typeFacts *typefacts.TypeFacts, fieldUsage string) metrics.MetricResult {
	sets := cohesion.EffectiveFieldSets(typeFacts, cohesion.FieldUsageMode(fieldUsage))

	return metrics.LCOM(
		cohesion.TotalFieldAccesses(sets),
		len(typeFacts.Fields),
		len(typeFacts.Methods),
	)
}
