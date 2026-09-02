package application

import (
	"testing"

	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
)

// White-box: per-method complexity and the derived AMC.
func TestComputeForTypeAMC(t *testing.T) {
	t.Parallel()

	tf := &typefacts.TypeFacts{Methods: []typefacts.MethodFacts{
		{Name: "A", Branches: typefacts.BranchStats{Ifs: 2}},            // CC 3
		{Name: "B", Branches: typefacts.BranchStats{}},                  // CC 1
		{Name: "C", Branches: typefacts.BranchStats{Fors: 1, Cases: 1}}, // CC 3
	}}
	got := ComputeForType(tf)

	want := []int{3, 1, 3}
	if len(got.MethodComplexities) != len(want) {
		t.Fatalf("complexities = %v", got.MethodComplexities)
	}

	for i := range want {
		if got.MethodComplexities[i] != want[i] {
			t.Errorf("method %d CC = %d, want %d", i, got.MethodComplexities[i], want[i])
		}
	}

	if !got.AMC.Applicable || got.AMC.Value != 7.0/3 {
		t.Fatalf("AMC = %+v, want 7/3", got.AMC)
	}
}

// White-box: a type with no methods has no AMC.
func TestComputeForTypeNoMethods(t *testing.T) {
	t.Parallel()

	got := ComputeForType(&typefacts.TypeFacts{})
	if len(got.MethodComplexities) != 0 {
		t.Errorf("complexities = %v, want empty", got.MethodComplexities)
	}

	if got.AMC.Applicable {
		t.Errorf("AMC should be n/a, got value %v", got.AMC.Value)
	}
}
