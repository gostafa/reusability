// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package metrics

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
)

type (
	metricSpec struct {
		name       string
		scope      MetricScope
		definition string
	}

	weightedComponent struct {
		component ReusabilityComponent
		weight    float64
	}
)

// AMC computes Average Method Complexity for a type:
//
//	AMC = Σ(method complexities) / methodCount
//
// Not applicable when the type has no methods.
func AMC(totalComplexity, methodCount int) MetricResult {
	spec := metricSpec{MetricAMC, ScopeType, DefinitionAMC}

	if methodCount == zero {
		return notApplicable(&spec, reasonNoMethods)
	}

	return applicable(&spec, float64(totalComplexity)/float64(methodCount))
}

// CBO computes type-level Coupling Between Objects:
//
//	CBO(t) = |ReferencedTypes(t)|
//
// referencedTypeCount is the number of distinct other analyzed named types
// the type references through its fields, method parameters, method returns,
// and embedded types (self-references excluded). Always applicable; the
// value may be 0.
func CBO(referencedTypeCount int) MetricResult {
	return applicable(
		&metricSpec{MetricCBO, ScopeType, DefinitionCBO},
		float64(referencedTypeCount),
	)
}

// Closure expands a selected display set to the full compute set: the
// transitive closure over metric dependencies, deduplicated, in a
// deterministic order. A metric computed only to satisfy a dependency is
// not rendered unless also selected.
func Closure(selected []string) []string {
	seen := make(map[string]bool, len(selected))

	for i := range selected {
		markClosure(seen, selected[i])
	}

	return orderedSeen(seen)
}

// CohesionComponent derives the cohesion component from an LCOM result:
//
//	cohesion = 1 − LCOM
//
// The component is dropped when LCOM is not applicable.
func CohesionComponent(lcom *MetricResult) ReusabilityComponent {
	if !lcom.Applicable {
		return ReusabilityComponent{Name: ComponentCohesion, Reason: lcom.Reason}
	}

	return ReusabilityComponent{Name: ComponentCohesion, Value: one - lcom.Value, Applicable: true}
}

// CouplingComponent derives the coupling component from a type's CBO using
// the saturating transform:
//
//	coupling = CBO / (CBO + 1)
//	component = 1 − coupling
//
// Always applicable.
func CouplingComponent(cbo int) ReusabilityComponent {
	coupling := float64(cbo) / (float64(cbo) + one)

	return ReusabilityComponent{Name: ComponentCoupling, Value: one - coupling, Applicable: true}
}

// DefaultReusabilityWeights returns the default component weights.
func DefaultReusabilityWeights() ReusabilityWeights {
	return ReusabilityWeights{
		Cohesion:      defaultWeightCohesion,
		Coupling:      defaultWeightCoupling,
		Testability:   defaultWeightCoupling,
		Documentation: defaultWeightDocs,
	}
}

// DocumentationComponent derives the documentation component:
//
//	documentation = documentedExportedMembers / exportedMembers
//
// The component is dropped when the type has no exported members.
func DocumentationComponent(documentedExportedMembers, exportedMembers int) ReusabilityComponent {
	if exportedMembers == zero {
		return ReusabilityComponent{
			Name:   ComponentDocumentation,
			Reason: "type has no exported members",
		}
	}

	return ReusabilityComponent{
		Name:       ComponentDocumentation,
		Value:      float64(documentedExportedMembers) / float64(exportedMembers),
		Applicable: true,
	}
}

// LCOM computes the matrix-density lack-of-cohesion metric:
//
//	LCOM = 1 − totalMethodFieldAccesses / (fieldCount × methodCount)
//
// totalMethodFieldAccesses is the number of 1-cells in the method-field
// matrix (each method contributes each distinct field it uses once). The
// result is in [0, 1]. Not applicable when fieldCount == 0 or
// methodCount == 0.
func LCOM(totalMethodFieldAccesses, fieldCount, methodCount int) MetricResult {
	spec := metricSpec{MetricLCOM, ScopeType, DefinitionLCOM}

	if fieldCount == zero {
		return notApplicable(&spec, "type has no fields")
	}

	if methodCount == zero {
		return notApplicable(&spec, reasonNoMethods)
	}

	density := float64(totalMethodFieldAccesses) / (float64(fieldCount) * float64(methodCount))

	return applicable(&spec, one-density)
}

// PackageMetricOrder is the fixed rendering order of package-level metrics.
// This linter reports no package-level metric, so the order is empty.
func PackageMetricOrder() []string {
	return nil
}

// ReportedMetricOrder is the single public metric this linter renders.
func ReportedMetricOrder() []string {
	return []string{MetricReusability}
}

// Reusability combines the four components into the Experimental Reusability
// Index:
//
//	RI = wc·cohesion + wk·(1 − coupling) + wt·testability + wd·documentation
//
// Components that are not applicable are dropped and the remaining weights
// are renormalized to sum to 1, keeping RI in [0, 1] and never yielding NaN.
// The result's Reason lists any dropped components. Not applicable only when
// no applicable component carries weight; the reason then spells out each
// dropped component with its own cause.
func Reusability(in *ReusabilityInput) MetricResult {
	inputs := weightedInputs(in)
	weightSum, dropped := sumWeights(inputs)
	names, details := droppedLabels(dropped)
	spec := metricSpec{MetricReusability, ScopeType, DefinitionReusability}

	if weightSum == zero {
		return reusabilityUnavailable(spec, inputs, dropped, details)
	}

	return reusabilityValue(spec, inputs, weightSum, names)
}

// Validate reports an error when any weight is negative or when the weights
// cannot be normalized (all zero).
func (weights *ReusabilityWeights) Validate() error {
	errs := [...]error{
		checkWeight(ComponentCohesion, weights.Cohesion),
		checkWeight(ComponentCoupling, weights.Coupling),
		checkWeight(ComponentTestability, weights.Testability),
		checkWeight(ComponentDocumentation, weights.Documentation),
	}

	for i := range errs {
		if errs[i] != nil {
			return errs[i]
		}
	}

	if weights.Cohesion+weights.Coupling+weights.Testability+weights.Documentation == zero {
		return errWeightsSumZero
	}

	return nil
}

func checkWeight(name string, value float64) error {
	if value < zero {
		return fmt.Errorf("%w: %q is %v", errNegativeWeight, name, value)
	}

	return nil
}

// TCC computes Tight Class Cohesion. A method pair is connected when the two
// methods share at least one field:
//
//	TCC = connectedPairs / totalPossiblePairs    = k(k−1)/2
//
// Not applicable when methodCount < 2 (totalPossiblePairs == 0).
func TCC(connectedPairs, methodCount int) MetricResult {
	spec := metricSpec{MetricTCC, ScopeType, DefinitionTCC}

	if methodCount < two {
		return notApplicable(&spec, "type has fewer than 2 methods")
	}

	totalPairs := methodCount * (methodCount - one) / two

	return applicable(&spec, float64(connectedPairs)/float64(totalPairs))
}

// TestabilityComponent derives the testability component from an AMC result:
//
//	testability = 1 / (1 + max(0, AMC − 1))
//
// 1 at AMC = 1, approaching 0 as AMC grows. The component is dropped when
// AMC is not applicable.
func TestabilityComponent(amc *MetricResult) ReusabilityComponent {
	if !amc.Applicable {
		return ReusabilityComponent{Name: ComponentTestability, Reason: amc.Reason}
	}

	return ReusabilityComponent{
		Name:       ComponentTestability,
		Value:      one / (one + max(zero, amc.Value-one)),
		Applicable: true,
	}
}

// TypeMetricOrder is the internal compute order of type-level metrics.
// AMC, LCOM, TCC, and CBO feed reusability and are never reported.
func TypeMetricOrder() []string {
	return []string{
		MetricAMC, MetricLCOM, MetricTCC,
		MetricCBO, MetricReusability,
	}
}

func applicable(spec *metricSpec, value float64) MetricResult {
	return MetricResult{
		Name:       spec.name,
		Scope:      spec.scope,
		Value:      value,
		Applicable: true,
		Definition: spec.definition,
	}
}

func appendOrdered(dst, order []string, seen map[string]bool) []string {
	for i := range order {
		if seen[order[i]] {
			dst = append(dst, order[i])
		}
	}

	return dst
}

func droppedLabels(dropped []ReusabilityComponent) (names, details []string) {
	slices.SortFunc(dropped, func(a, b ReusabilityComponent) int {
		return cmp.Compare(a.Name, b.Name)
	})

	names = make([]string, zero, len(dropped))
	details = make([]string, zero, len(dropped))

	for i := range dropped {
		names = append(names, dropped[i].Name)

		detail := dropped[i].Name

		if dropped[i].Reason != "" {
			detail = dropped[i].Name + " (" + dropped[i].Reason + ")"
		}

		details = append(details, detail)
	}

	return names, details
}

func markClosure(seen map[string]bool, name string) {
	if seen[name] {
		return
	}

	seen[name] = true

	deps := metricDependencies(name)

	for i := range deps {
		markClosure(seen, deps[i])
	}
}

func metricDependencies(name string) []string {
	if name == MetricReusability {
		return []string{MetricLCOM, MetricAMC, MetricCBO}
	}

	return nil
}

func notApplicable(spec *metricSpec, reason string) MetricResult {
	return MetricResult{
		Name:       spec.name,
		Scope:      spec.scope,
		Applicable: false,
		Reason:     reason,
		Definition: spec.definition,
	}
}

func orderedSeen(seen map[string]bool) []string {
	closure := make([]string, zero, len(seen))

	closure = appendOrdered(closure, TypeMetricOrder(), seen)
	closure = appendOrdered(closure, PackageMetricOrder(), seen)

	return closure
}

func reusabilityUnavailable(
	spec metricSpec,
	inputs []weightedComponent,
	dropped []ReusabilityComponent,
	details []string,
) MetricResult {
	if len(dropped) == len(inputs) {
		return notApplicable(
			&spec,
			"every component dropped: "+strings.Join(details, listSep),
		)
	}

	return notApplicable(
		&spec,
		"the applicable components have zero total weight; dropped: "+strings.Join(
			details,
			listSep,
		),
	)
}

func reusabilityValue(
	spec metricSpec,
	inputs []weightedComponent,
	weightSum float64,
	names []string,
) MetricResult {
	var value float64

	for i := range inputs {
		in := &inputs[i]

		if in.component.Applicable {
			value += in.weight / weightSum * in.component.Value
		}
	}

	result := applicable(&spec, value)

	if len(names) > zero {
		result.Reason = "dropped components: " + strings.Join(names, listSep)
	}

	return result
}

func sumWeights(inputs []weightedComponent) (weightSum float64, dropped []ReusabilityComponent) {
	for i := range inputs {
		if inputs[i].component.Applicable {
			weightSum += inputs[i].weight

			continue
		}

		dropped = append(dropped, inputs[i].component)
	}

	return weightSum, dropped
}

func weightedInputs(in *ReusabilityInput) []weightedComponent {
	return []weightedComponent{
		{in.Cohesion, in.Weights.Cohesion},
		{in.Coupling, in.Weights.Coupling},
		{in.Testability, in.Weights.Testability},
		{in.Documentation, in.Weights.Documentation},
	}
}
