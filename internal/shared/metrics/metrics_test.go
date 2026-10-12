// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package metrics

import (
	"strings"
	"testing"
)

func TestMetricDefinitionsAndClosure(t *testing.T) {
	t.Parallel()
	if got := ReportedMetricOrder(); len(got) != 1 || got[0] != MetricReusability {
		t.Fatalf("reported metrics = %v", got)
	}
	for _, result := range []MetricResult{AMC(0, 0), LCOM(0, 0, 1), LCOM(0, 1, 0), TCC(0, 1)} {
		if result.Applicable || result.Reason == "" {
			t.Fatalf("missing n/a reason: %+v", result)
		}
	}
	if got := CBO(3); got.Value != 3 || !got.Applicable {
		t.Fatalf("CBO = %+v", got)
	}
	if got := TCC(1, 2); got.Value != 1 {
		t.Fatalf("TCC = %+v", got)
	}
	got := Closure([]string{MetricReusability, MetricLCOM, MetricReusability})
	if strings.Join(got, ",") != "amc,lcom,cbo,reusability" {
		t.Fatalf("closure = %v", got)
	}
	if len(PackageMetricOrder()) != 0 {
		t.Fatal("unexpected package metrics")
	}
	weights := ReusabilityWeights{}
	if weights.Validate() == nil {
		t.Fatal("zero weights accepted")
	}
}

func TestDroppedComponentsAndZeroWeights(t *testing.T) {
	t.Parallel()
	na := AMC(0, 0)
	cohesion, testability := CohesionComponent(&na), TestabilityComponent(&na)
	docs := DocumentationComponent(0, 0)
	input := ReusabilityInput{
		Cohesion:      cohesion,
		Testability:   testability,
		Documentation: docs,
		Weights:       DefaultReusabilityWeights(),
	}
	if got := Reusability(
		&input,
	); got.Applicable ||
		!strings.Contains(got.Reason, "every component dropped") {
		t.Fatalf("all dropped = %+v", got)
	}
	input.Coupling = CouplingComponent(0)
	if got := Reusability(
		&input,
	); !got.Applicable || got.Value != 1 ||
		!strings.Contains(got.Reason, "dropped components") {
		t.Fatalf("partial = %+v", got)
	}
	input.Weights = ReusabilityWeights{}
	if got := Reusability(
		&input,
	); got.Applicable ||
		!strings.Contains(got.Reason, "zero total weight") {
		t.Fatalf("zero weight = %+v", got)
	}
}

// Black-box: the reusability index composes the four components into [0, 1].
func TestReusabilityComposition(t *testing.T) {
	t.Parallel()

	lcom := LCOM(4, 2, 2)
	amc := AMC(2, 2)
	r := Reusability(&ReusabilityInput{
		Cohesion:      CohesionComponent(&lcom),
		Coupling:      CouplingComponent(1),
		Testability:   TestabilityComponent(&amc),
		Documentation: DocumentationComponent(2, 2),
		Weights:       DefaultReusabilityWeights(),
	})
	if !r.Applicable {
		t.Fatalf("reusability not applicable: %s", r.Reason)
	}

	if r.Value < 0 || r.Value > 1 {
		t.Errorf("reusability %v out of [0,1]", r.Value)
	}
}

// Black-box: an all-zero weight set is rejected by validation.
func TestWeightsValidate(t *testing.T) {
	t.Parallel()

	defaults := DefaultReusabilityWeights()
	err := defaults.Validate()
	if err != nil {
		t.Errorf("defaults should validate: %v", err)
	}

	bad := ReusabilityWeights{Cohesion: -1}
	err = bad.Validate()
	if err == nil {
		t.Error("negative weight should fail validation")
	}
}
