package domain_test

import (
	"testing"

	cohesion "github.com/gostafa/reusability/internal/features/cohesion/domain"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/bitset"
)

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
			{Name: "M2", FieldsUsed: fs(0)}, // shares field 0
		},
	}

	sets := cohesion.EffectiveFieldSets(tf, false)

	if got := cohesion.CountSharingPairs(sets, len(tf.Fields)); got != 1 {
		t.Fatalf("sharing pairs = %d, want 1", got)
	}

	if got := cohesion.TotalFieldAccesses(sets); got != 2 {
		t.Errorf("total accesses = %d, want 2", got)
	}
}
