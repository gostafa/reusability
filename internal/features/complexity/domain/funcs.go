// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
)

const (
	zero = 0
	one  = 1
)

// Cyclomatic computes a method's cyclomatic complexity: base 1, incremented
// for each if, for, range, non-default case, select communication clause,
// &&, and ||.
func Cyclomatic(branches *typefacts.BranchStats) int {
	return one + branches.Ifs + branches.Fors + branches.Ranges + branches.Cases +
		branches.SelectComms + branches.LogicalOps
}
