package domain_test

import (
	"testing"

	complexity "github.com/gostafa/reusability/internal/features/complexity/domain"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
)

// Black-box: the exported contract as a consumer of the feature sees it.
func TestCyclomaticContract(t *testing.T) {
	t.Parallel()
	// A method with no branches is minimally complex.
	if got := complexity.Cyclomatic(typefacts.BranchStats{}); got != 1 {
		t.Fatalf("straight-line = %d, want 1", got)
	}
	// A realistic method: two guards, a loop, and one && — 1 + 2 + 1 + 1.
	realistic := typefacts.BranchStats{Ifs: 2, Fors: 1, LogicalOps: 1}
	if got := complexity.Cyclomatic(realistic); got != 5 {
		t.Fatalf("realistic method = %d, want 5", got)
	}
}
