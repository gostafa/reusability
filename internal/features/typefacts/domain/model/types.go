// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

// Package model provides reusable analysis utilities.
package model

import "github.com/gostafa/reusability/internal/shared/bitset"

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

type (
	// Position is a source location for a declaration.
	Position = struct {
		// File is the source file path, relative when possible.
		File string
		// Line is the 1-based source line.
		Line int
		// Column is the 1-based source column.
		Column int
	}

	// FieldFacts describes one struct field.
	FieldFacts = struct {
		// Name is the field name (the type name for embedded fields).
		Name string
		// Exported reports whether the field name is exported.
		Exported bool
		// Embedded marks an embedded (anonymous) field.
		Embedded bool
	}

	// DeclarationFacts is the common name/position/export metadata.
	DeclarationFacts = struct {
		// Name is the declaration identifier.
		Name string
		// Pos is the source location of the declaration.
		Pos Position
		// Exported reports whether the declaration name is exported.
		Exported bool
	}
)
