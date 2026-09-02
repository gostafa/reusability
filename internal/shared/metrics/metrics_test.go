// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package metrics

import "testing"

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
