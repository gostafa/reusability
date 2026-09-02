package application

import (
	"testing"

	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/bitset"
)

func fieldSet(size int, idx ...int) bitset.FieldSet {
	s := bitset.NewFieldSet(size)
	for _, i := range idx {
		s.Set(i)
	}

	return s
}

// White-box: two methods over disjoint fields have no connected pairs.
func TestComputeForTypeDisjoint(t *testing.T) {
	t.Parallel()

	tf := &typefacts.TypeFacts{
		Fields: []typefacts.FieldFacts{{Name: "a"}, {Name: "b"}},
		Methods: []typefacts.MethodFacts{
			{Name: "M1", FieldsUsed: fieldSet(2, 0)},
			{Name: "M2", FieldsUsed: fieldSet(2, 1)},
		},
	}
	got := ComputeForType(tf, false)

	// No connected pairs → TCC = 0.
	if !got.TCC.Applicable || got.TCC.Value != 0 {
		t.Errorf("TCC = %+v, want 0", got.TCC)
	}
}

// White-box: transitive mode propagates field usage through sibling calls,
// making the two methods share and become cohesive.
func TestComputeForTypeTransitive(t *testing.T) {
	t.Parallel()

	tf := &typefacts.TypeFacts{
		Fields: []typefacts.FieldFacts{{Name: "a"}, {Name: "b"}},
		Methods: []typefacts.MethodFacts{
			{Name: "M1", FieldsUsed: fieldSet(2, 0), CalledSiblings: []int{1}},
			{Name: "M2", FieldsUsed: fieldSet(2, 1)},
		},
	}
	direct := ComputeForType(tf, false)
	transitive := ComputeForType(tf, true)

	if direct.TCC.Value != 0 {
		t.Errorf("direct TCC = %v, want 0", direct.TCC.Value)
	}

	if transitive.TCC.Value != 1 {
		t.Errorf("transitive TCC = %v, want 1 (M1 now shares via M2)", transitive.TCC.Value)
	}
}
