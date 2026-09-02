// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"testing"

	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/bitset"
)

func fieldSet(size int, indices ...int) bitset.FieldSet {
	s := bitset.NewFieldSet(size)
	for _, i := range indices {
		s.Set(i)
	}

	return s
}

func TestCountSharingPairsSmallPath(t *testing.T) {
	sets := []bitset.FieldSet{
		fieldSet(3, 0),
		fieldSet(3, 0, 1),
		fieldSet(3, 2),
	}

	if got := CountSharingPairs(sets, 3); got != 1 {
		t.Fatalf("sharing pairs = %d, want 1", got)
	}
}

func TestCountSharingPairsGeneralPath(t *testing.T) {
	sets := []bitset.FieldSet{
		fieldSet(70, 69),
		fieldSet(70, 69, 1),
		fieldSet(70, 5),
	}

	if got := CountSharingPairs(sets, 70); got != 1 {
		t.Fatalf("sharing pairs = %d, want 1", got)
	}
}

func TestCountSharingPairsDegenerate(t *testing.T) {
	if got := CountSharingPairs(nil, 0); got != 0 {
		t.Fatalf("sharing pairs = %d, want 0", got)
	}

	if got := CountSharingPairs([]bitset.FieldSet{{}, {}}, 0); got != 0 {
		t.Fatalf("sharing pairs = %d, want 0", got)
	}
}

func TestEffectiveFieldSetsTransitive(t *testing.T) {
	facts := &typefacts.TypeFacts{
		Fields: []typefacts.FieldFacts{{Name: "x"}, {Name: "y"}},
		Methods: []typefacts.MethodFacts{
			{Name: "a", FieldsUsed: fieldSet(2, 0), CalledSiblings: []int{1}},
			{Name: "b", FieldsUsed: fieldSet(2), CalledSiblings: []int{2}},
			{Name: "c", FieldsUsed: fieldSet(2, 1)},
		},
	}

	direct := EffectiveFieldSets(facts, FieldUsageDirect)
	if bitset.Count(direct[0]) != 1 || bitset.Count(direct[1]) != 0 {
		t.Fatalf("direct sets changed: %v", direct)
	}

	transitive := EffectiveFieldSets(facts, FieldUsageTransitive)
	if !bitset.Contains(transitive[0], 1) {
		t.Fatal("a should reach c's field through b (fixpoint)")
	}

	if bitset.Count(transitive[1]) != 1 || !bitset.Contains(transitive[1], 1) {
		t.Fatal("b should absorb c's field")
	}

	if bitset.Contains(facts.Methods[0].FieldsUsed, 1) {
		t.Fatal("transitive mode mutated the extracted facts")
	}
}

func TestTotalFieldAccesses(t *testing.T) {
	sets := []bitset.FieldSet{fieldSet(3, 0, 1), fieldSet(3, 1), {}}
	if got := TotalFieldAccesses(sets); got != 3 {
		t.Fatalf("total = %d, want 3", got)
	}
}

// Black-box: field-set derivation and sharing-pair counting from the exported API.
func TestFieldSetsAndPairs(t *testing.T) {
	t.Parallel()

	fs := func(indices ...int) bitset.FieldSet {
		s := bitset.NewFieldSet(2)
		for _, i := range indices {
			s.Set(i)
		}

		return s
	}
	tf := &typefacts.TypeFacts{
		Fields: []typefacts.FieldFacts{{Name: "a"}, {Name: "b"}},
		Methods: []typefacts.MethodFacts{
			{Name: "M1", FieldsUsed: fs(0)},
			{Name: "M2", FieldsUsed: fs(0)},
		},
	}

	sets := EffectiveFieldSets(tf, FieldUsageDirect)

	if got := CountSharingPairs(sets, len(tf.Fields)); got != 1 {
		t.Fatalf("sharing pairs = %d, want 1", got)
	}

	if got := TotalFieldAccesses(sets); got != 2 {
		t.Errorf("total accesses = %d, want 2", got)
	}
}
