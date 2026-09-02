// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"testing"

	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
)

// White-box: the cyclomatic formula over branch facts.
func TestCyclomatic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		b    typefacts.BranchStats
		want int
	}{
		{"straight line is base 1", typefacts.BranchStats{}, 1},
		{"single if", typefacts.BranchStats{Ifs: 1}, 2},
		{"each branch kind adds one", typefacts.BranchStats{
			Ifs: 1, Fors: 1, Ranges: 1, Cases: 1, SelectComms: 1, LogicalOps: 1,
		}, 7},
		{"counts accumulate", typefacts.BranchStats{
			Ifs: 3, Fors: 2, Ranges: 1, Cases: 4, SelectComms: 0, LogicalOps: 5,
		}, 1 + 3 + 2 + 1 + 4 + 0 + 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := Cyclomatic(&tt.b); got != tt.want {
				t.Errorf("Cyclomatic(%+v) = %d, want %d", tt.b, got, tt.want)
			}
		})
	}
}

// Black-box: the exported contract as a consumer of the feature sees it.
func TestCyclomaticContract(t *testing.T) {
	t.Parallel()

	empty := typefacts.BranchStats{}
	if got := Cyclomatic(&empty); got != 1 {
		t.Fatalf("straight-line = %d, want 1", got)
	}

	realistic := typefacts.BranchStats{Ifs: 2, Fors: 1, LogicalOps: 1}
	if got := Cyclomatic(&realistic); got != 5 {
		t.Fatalf("realistic method = %d, want 5", got)
	}
}
