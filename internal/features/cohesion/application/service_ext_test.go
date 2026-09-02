package application_test

import (
	"testing"

	cohesion "github.com/gostafa/reusability/internal/features/cohesion/application"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/bitset"
)

// Black-box: two methods sharing a field form one connected pair, giving
// full TCC.
func TestComputeForTypeCohesive(t *testing.T) {
	t.Parallel()

	shared := bitset.NewFieldSet(1)
	shared.Set(0)
	tf := &typefacts.TypeFacts{
		Fields: []typefacts.FieldFacts{{Name: "a"}},
		Methods: []typefacts.MethodFacts{
			{Name: "Get", FieldsUsed: cloneSet(shared)},
			{Name: "Set", FieldsUsed: cloneSet(shared)},
		},
	}

	got := cohesion.ComputeForType(tf, false)
	if !got.TCC.Applicable || got.TCC.Value != 1 {
		t.Errorf("TCC = %+v, want 1", got.TCC)
	}

}

func cloneSet(s bitset.FieldSet) bitset.FieldSet {
	return bitset.Clone(s)
}
