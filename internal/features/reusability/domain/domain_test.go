// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"testing"

	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/bitset"
	"github.com/gostafa/reusability/internal/shared/metrics"
)

func fieldSet(size int, idx ...int) bitset.FieldSet {
	s := bitset.NewFieldSet(size)
	for _, i := range idx {
		s.Set(i)
	}

	return s
}

// White-box: Compute derives CBO from the referenced-types fact and folds
// the four components into the index.
func TestComputeDerivesCBOAndIndex(t *testing.T) {
	t.Parallel()

	tf := &typefacts.TypeFacts{
		ReferencedTypeIDs: []int{2, 5, 9},
		Fields:            []typefacts.FieldFacts{{Name: "a"}, {Name: "b"}, {Name: "c"}},
		Methods: []typefacts.MethodFacts{
			{Name: "M1", FieldsUsed: fieldSet(3, 0), Branches: typefacts.BranchStats{Ifs: 1}},
			{Name: "M2", FieldsUsed: fieldSet(3, 1), Branches: typefacts.BranchStats{Ifs: 1}},
			{Name: "M3", FieldsUsed: fieldSet(3, 0, 1), Branches: typefacts.BranchStats{}},
		},
		ExportedMembers:           4,
		DocumentedExportedMembers: 3,
	}

	weights := metrics.DefaultReusabilityWeights()
	got := Compute(tf, &weights, "direct")

	if got.CBO != metrics.CBO(len(tf.ReferencedTypeIDs)) {
		t.Errorf("CBO = %+v, want %+v", got.CBO, metrics.CBO(3))
	}

	if !got.Reusability.Applicable {
		t.Fatalf("reusability not applicable: %s", got.Reusability.Reason)
	}
}

// White-box: when the type has no methods/fields, cohesion and testability
// drop and the index renormalizes.
func TestComputeDropsNotApplicableComponents(t *testing.T) {
	t.Parallel()

	tf := &typefacts.TypeFacts{ExportedMembers: 2, DocumentedExportedMembers: 1}

	weights := metrics.DefaultReusabilityWeights()
	got := Compute(tf, &weights, "direct")

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
	fields := []typefacts.FieldFacts{{Name: "a"}, {Name: "b"}}
	methods := []typefacts.MethodFacts{
		{Name: "M1", FieldsUsed: fieldSet(2, 0, 1), Branches: typefacts.BranchStats{Ifs: 1}},
		{Name: "M2", FieldsUsed: fieldSet(2, 0, 1), Branches: typefacts.BranchStats{Ifs: 1}},
	}

	good := Compute(
		&typefacts.TypeFacts{
			Fields: fields, Methods: methods,
			ExportedMembers: 4, DocumentedExportedMembers: 4,
		},
		&weights,
		"direct",
	)
	poor := Compute(
		&typefacts.TypeFacts{
			ReferencedTypeIDs: []int{1, 2, 3, 4, 5},
			Fields:            fields,
			Methods:           methods,
			ExportedMembers:   4, DocumentedExportedMembers: 0,
		},
		&weights,
		"direct",
	)

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
