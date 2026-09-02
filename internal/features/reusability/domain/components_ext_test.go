package domain_test

import (
	"testing"

	reusability "github.com/gostafa/reusability/internal/features/reusability/domain"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
)

// Black-box: a fully-documented, cohesive, uncoupled type scores higher than
// an undocumented, coupled one under the same weights.
func TestComputeIndexRewardsQuality(t *testing.T) {
	t.Parallel()

	weights := metrics.DefaultReusabilityWeights()
	amc := metrics.AMC(2, 2)
	lcom := metrics.LCOM(4, 2, 2)

	good := reusability.Compute(&typefacts.TypeFacts{
		ExportedMembers: 4, DocumentedExportedMembers: 4,
	}, amc, lcom, weights)
	poor := reusability.Compute(&typefacts.TypeFacts{
		ReferencedTypeIDs: []int{1, 2, 3, 4, 5},
		ExportedMembers:   4, DocumentedExportedMembers: 0,
	}, amc, lcom, weights)

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
