package metrics

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// MetricReusability is the metric name for the Experimental Reusability Index.
const MetricReusability = "reusability"

// DefinitionReusability versions the Reusability formula. The metric is
// experimental.
const DefinitionReusability = "reusability/reusability-v1"

// Component names used in ReusabilityWeights and dropped-component reasons.
const (
	ComponentCohesion      = "cohesion"
	ComponentCoupling      = "coupling"
	ComponentTestability   = "testability"
	ComponentDocumentation = "documentation"
)

// ReusabilityWeights holds the weight of each reusability component.
type ReusabilityWeights struct {
	Cohesion      float64 // weight of the cohesion component (from LCOM)
	Coupling      float64 // weight of the coupling component (from CBO)
	Testability   float64 // weight of the testability component (from AMC)
	Documentation float64 // weight of the documentation component
}

// DefaultReusabilityWeights returns the default component weights.
func DefaultReusabilityWeights() ReusabilityWeights {
	return ReusabilityWeights{
		Cohesion:      0.35,
		Coupling:      0.25,
		Testability:   0.25,
		Documentation: 0.15,
	}
}

// Validate reports an error when any weight is negative or when the weights
// cannot be normalized (all zero).
func (w ReusabilityWeights) Validate() error {
	for _, c := range []struct {
		name  string
		value float64
	}{
		{ComponentCohesion, w.Cohesion},
		{ComponentCoupling, w.Coupling},
		{ComponentTestability, w.Testability},
		{ComponentDocumentation, w.Documentation},
	} {
		if c.value < 0 {
			return fmt.Errorf("reusability weight %q is negative: %v", c.name, c.value)
		}
	}

	if w.Cohesion+w.Coupling+w.Testability+w.Documentation == 0 {
		return errors.New("reusability weights sum to zero and cannot be normalized")
	}

	return nil
}

// ReusabilityComponent is one normalized [0, 1] input to the reusability
// formula. Value already carries the component's contribution orientation
// (e.g. the coupling component is 1 − saturatedCoupling).
type ReusabilityComponent struct {
	Name       string  // component name, e.g. "cohesion"
	Value      float64 // normalized contribution; meaningless unless Applicable
	Applicable bool    // whether the component participates in the index
	Reason     string  // why the component is dropped, when not applicable
}

// CohesionComponent derives the cohesion component from an LCOM result:
//
//	cohesion = 1 − LCOM
//
// The component is dropped when LCOM is not applicable.
func CohesionComponent(lcom MetricResult) ReusabilityComponent {
	if !lcom.Applicable {
		return ReusabilityComponent{Name: ComponentCohesion, Reason: lcom.Reason}
	}

	return ReusabilityComponent{Name: ComponentCohesion, Value: 1 - lcom.Value, Applicable: true}
}

// CouplingComponent derives the coupling component from a type's CBO using
// the saturating transform:
//
//	coupling = CBO / (CBO + 1)
//	component = 1 − coupling
//
// Always applicable.
func CouplingComponent(cbo int) ReusabilityComponent {
	coupling := float64(cbo) / (float64(cbo) + 1)

	return ReusabilityComponent{Name: ComponentCoupling, Value: 1 - coupling, Applicable: true}
}

// TestabilityComponent derives the testability component from an AMC result:
//
//	testability = 1 / (1 + max(0, AMC − 1))
//
// 1 at AMC = 1, approaching 0 as AMC grows. The component is dropped when
// AMC is not applicable.
func TestabilityComponent(amc MetricResult) ReusabilityComponent {
	if !amc.Applicable {
		return ReusabilityComponent{Name: ComponentTestability, Reason: amc.Reason}
	}

	return ReusabilityComponent{
		Name:       ComponentTestability,
		Value:      1 / (1 + max(0, amc.Value-1)),
		Applicable: true,
	}
}

// DocumentationComponent derives the documentation component:
//
//	documentation = documentedExportedMembers / exportedMembers
//
// The component is dropped when the type has no exported members.
func DocumentationComponent(documentedExportedMembers, exportedMembers int) ReusabilityComponent {
	if exportedMembers == 0 {
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
func Reusability(
	cohesion, coupling, testability, documentation ReusabilityComponent,
	weights ReusabilityWeights,
) MetricResult {
	inputs := []struct {
		component ReusabilityComponent
		weight    float64
	}{
		{cohesion, weights.Cohesion},
		{coupling, weights.Coupling},
		{testability, weights.Testability},
		{documentation, weights.Documentation},
	}

	var (
		weightSum float64
		dropped   []ReusabilityComponent
	)

	for _, in := range inputs {
		if in.component.Applicable {
			weightSum += in.weight
		} else {
			dropped = append(dropped, in.component)
		}
	}

	sort.Slice(dropped, func(i, j int) bool { return dropped[i].Name < dropped[j].Name })
	names := make([]string, len(dropped))

	details := make([]string, len(dropped))
	for i, c := range dropped {
		names[i] = c.Name

		details[i] = c.Name
		if c.Reason != "" {
			details[i] = c.Name + " (" + c.Reason + ")"
		}
	}

	if weightSum == 0 {
		if len(dropped) == len(inputs) {
			return notApplicable(MetricReusability, ScopeType, DefinitionReusability,
				"every component dropped: "+strings.Join(details, ", "))
		}

		return notApplicable(
			MetricReusability,
			ScopeType,
			DefinitionReusability,
			"the applicable components have zero total weight; dropped: "+strings.Join(
				details,
				", ",
			),
		)
	}

	var value float64

	for _, in := range inputs {
		if in.component.Applicable {
			value += in.weight / weightSum * in.component.Value
		}
	}

	result := applicable(MetricReusability, ScopeType, DefinitionReusability, value)
	if len(dropped) > 0 {
		result.Reason = "dropped components: " + strings.Join(names, ", ")
	}

	return result
}
