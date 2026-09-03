// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package model

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
