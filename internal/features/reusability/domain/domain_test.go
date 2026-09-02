// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"testing"

	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
)

// White-box: Compute derives CBO from the referenced-types fact and folds
// the four components into the index.
func TestComputeDerivesCBOAndIndex(t *testing.T) {
	t.Parallel()

	tf := &typefacts.TypeFacts{
		ReferencedTypeIDs:         []int{2, 5, 9},
		ExportedMembers:           4,
		DocumentedExportedMembers: 3,
	}
	amc := metrics.AMC(6, 3)
	lcom := metrics.LCOM(2, 3, 3)
	weights := metrics.DefaultReusabilityWeights()

	got := Compute(&ComputeInput{Type: tf, AMC: amc, LCOM: lcom, Weights: weights})

	if got.CBO != metrics.CBO(len(tf.ReferencedTypeIDs)) {
		t.Errorf("CBO = %+v, want %+v", got.CBO, metrics.CBO(3))
	}

	if !got.Reusability.Applicable {
		t.Fatalf("reusability not applicable: %s", got.Reusability.Reason)
	}
}

// White-box: when the upstream cohesion/testability inputs are not
// applicable, their components are dropped and the index renormalizes.
func TestComputeDropsNotApplicableComponents(t *testing.T) {
	t.Parallel()

	tf := &typefacts.TypeFacts{ExportedMembers: 2, DocumentedExportedMembers: 1}
	amc := metrics.AMC(0, 0)
	lcom := metrics.LCOM(0, 0, 1)

	got := Compute(&ComputeInput{
		Type: tf, AMC: amc, LCOM: lcom, Weights: metrics.DefaultReusabilityWeights(),
	})

	if got.CBO != metrics.CBO(0) {
		t.Errorf("CBO = %+v, want %+v", got.CBO, metrics.CBO(0))
	}

	if got.Reusability.Applicable && got.Reusability.Reason == "" {
		t.Error("dropped-component index should record which components were dropped")
	}
}

// Black-box: a fully-documented, cohesive, uncoupled type scores higher than
// an undocumented, coupled one under the same weights.
func TestComputeIndexRewardsQuality(t *testing.T) {
	t.Parallel()

	weights := metrics.DefaultReusabilityWeights()
	amc := metrics.AMC(2, 2)
	lcom := metrics.LCOM(4, 2, 2)

	good := Compute(&ComputeInput{
		Type: &typefacts.TypeFacts{ExportedMembers: 4, DocumentedExportedMembers: 4},
		AMC:  amc, LCOM: lcom, Weights: weights,
	})
	poor := Compute(&ComputeInput{
		Type: &typefacts.TypeFacts{
			ReferencedTypeIDs: []int{1, 2, 3, 4, 5},
			ExportedMembers:   4, DocumentedExportedMembers: 0,
		},
		AMC: amc, LCOM: lcom, Weights: weights,
	})

	if !good.Reusability.Applicable || !poor.Reusability.Applicable {
		t.Fatal("both indices should be applicable")
	}

	if good.Reusability.Value <= poor.Reusability.Value {
		t.Errorf("documented/uncoupled %.3f should exceed undocumented/coupled %.3f",
			good.Reusability.Value, poor.Reusability.Value)
	}

	if poor.CBO.Value != 5 {
		t.Errorf("coupled type CBO = %v, want 5", poor.CBO.Value)
	}
}
