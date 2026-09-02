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
	amc := metrics.AMC(6, 3)      // applicable
	lcom := metrics.LCOM(2, 3, 3) // applicable
	weights := metrics.DefaultReusabilityWeights()

	got := Compute(tf, amc, lcom, weights)

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
	amc := metrics.AMC(0, 0)      // not applicable (no methods)
	lcom := metrics.LCOM(0, 0, 1) // not applicable

	got := Compute(tf, amc, lcom, metrics.DefaultReusabilityWeights())

	if got.CBO != metrics.CBO(0) {
		t.Errorf("CBO = %+v, want %+v", got.CBO, metrics.CBO(0))
	}

	if got.Reusability.Applicable && got.Reusability.Reason == "" {
		t.Error("dropped-component index should record which components were dropped")
	}
}
