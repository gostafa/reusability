package domain

import typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"

// Cyclomatic computes a method's cyclomatic complexity: base 1, incremented
// for each if, for, range, non-default case, select communication clause,
// &&, and ||.
func Cyclomatic(b typefacts.BranchStats) int {
	return 1 + b.Ifs + b.Fors + b.Ranges + b.Cases + b.SelectComms + b.LogicalOps
}
