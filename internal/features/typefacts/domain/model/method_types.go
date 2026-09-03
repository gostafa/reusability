// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package model

import (
	"github.com/gostafa/reusability/internal/shared/bitset"
)

type (
	// FunctionFacts describes a package-level function.
	FunctionFacts = struct {
		// Name is the function identifier.
		Name string
		// Pos is the source location of the function.
		Pos Position
		// Branches counts control-flow constructs for cyclomatic complexity.
		Branches BranchStats
		// Lines is the function body line count.
		Lines int
		// Exported reports whether the function name is exported.
		Exported bool
	}

	// MethodFacts describes one method of a named type.
	MethodFacts = struct {
		Name           string
		FieldsUsed     bitset.FieldSet
		CalledSiblings []int
		Pos            Position
		Branches       BranchStats
		Lines          int
		Exported       bool
	}

	// BranchStats counts control-flow constructs that raise cyclomatic complexity.
	BranchStats = struct {
		Ifs         int // if statements
		Fors        int // for loops (all, including conditionless "for {}")
		Ranges      int // range loops
		Cases       int // non-default switch and type-switch cases
		SelectComms int // select communication clauses (default excluded)
		LogicalOps  int // && and ||
	}
)
