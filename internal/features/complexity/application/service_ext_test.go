package application_test

import (
	"testing"

	complexity "github.com/gostafa/reusability/internal/features/complexity/application"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
)

// Black-box: the feature entry point turns branch facts into AMC.
func TestComputeForTypeContract(t *testing.T) {
	t.Parallel()

	tf := &typefacts.TypeFacts{Methods: []typefacts.MethodFacts{
		{Name: "Do", Branches: typefacts.BranchStats{Ifs: 1, LogicalOps: 1}}, // CC 3
		{Name: "Noop"}, // CC 1
	}}

	got := complexity.ComputeForType(tf)
	if !got.AMC.Applicable || got.AMC.Value != 2 {
		t.Fatalf("AMC = %+v, want 2", got.AMC)
	}
}
