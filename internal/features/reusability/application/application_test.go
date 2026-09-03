// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

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

// White-box: weight defaulting and validation at construction.
func TestNewServiceDefaultsAndValidation(t *testing.T) {
	t.Parallel()

	if _, err := NewService(&metrics.ReusabilityWeights{}); err != nil {
		t.Fatalf("zero weights should select defaults, got %v", err)
	}

	if _, err := NewService(&metrics.ReusabilityWeights{Cohesion: -1, Coupling: 1}); err == nil {
		t.Fatal("negative weight accepted")
	}
}

// White-box: the service delegates to the domain formula.
func TestServiceComputeForType(t *testing.T) {
	t.Parallel()

	svc, err := NewService(&metrics.ReusabilityWeights{})
	if err != nil {
		t.Fatal(err)
	}

	tf := &typefacts.TypeFacts{
		ReferencedTypeIDs: []int{1, 2},
		Fields:            []typefacts.FieldFacts{{Name: "a"}, {Name: "b"}},
		Methods: []typefacts.MethodFacts{
			{Name: "M1", FieldsUsed: fieldSet(2, 0, 1), Branches: typefacts.BranchStats{Ifs: 1}},
			{Name: "M2", FieldsUsed: fieldSet(2, 0, 1), Branches: typefacts.BranchStats{}},
		},
		ExportedMembers:           2,
		DocumentedExportedMembers: 2,
	}

	got := svc(tf, "direct")
	if got.CBO != metrics.CBO(2) {
		t.Errorf("CBO = %+v, want %+v", got.CBO, metrics.CBO(2))
	}

	if !got.Reusability.Applicable {
		t.Errorf("reusability not applicable: %s", got.Reusability.Reason)
	}
}

// Black-box: constructing with explicit weights and evaluating a type.
func TestServiceEndToEnd(t *testing.T) {
	t.Parallel()

	svc, err := NewService(&metrics.ReusabilityWeights{
		Cohesion: 0.4, Coupling: 0.3, Testability: 0.2, Documentation: 0.1,
	})
	if err != nil {
		t.Fatal(err)
	}

	tf := &typefacts.TypeFacts{
		ReferencedTypeIDs: []int{1, 2, 3},
		Fields:            []typefacts.FieldFacts{{Name: "a"}, {Name: "b"}},
		Methods: []typefacts.MethodFacts{
			{Name: "M1", FieldsUsed: fieldSet(2, 0), Branches: typefacts.BranchStats{Ifs: 3}},
			{Name: "M2", FieldsUsed: fieldSet(2, 1), Branches: typefacts.BranchStats{}},
		},
		ExportedMembers:           1,
		DocumentedExportedMembers: 1,
	}

	got := svc(tf, "direct")
	if got.CBO.Value != 3 {
		t.Errorf("CBO = %v, want 3", got.CBO.Value)
	}

	if !got.Reusability.Applicable {
		t.Errorf("reusability not applicable: %s", got.Reusability.Reason)
	}
}

// Black-box: a negative weight is rejected at construction.
func TestNewServiceRejectsNegativeWeight(t *testing.T) {
	t.Parallel()

	if _, err := NewService(
		&metrics.ReusabilityWeights{Cohesion: -0.5, Coupling: 1},
	); err == nil {
		t.Fatal("negative weight accepted")
	}
}
