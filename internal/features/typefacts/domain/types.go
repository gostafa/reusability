// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	factmodel "github.com/gostafa/reusability/internal/features/typefacts/domain/model"
)

type (
	// PackageExtract is one package's raw extraction before ID assignment.
	PackageExtract struct {
		Path     string
		Imports  []string
		Types    []TypeExtract
		InModule bool
	}

	// TypeExtract is one named type's raw extraction before ID assignment.
	TypeExtract struct {
		Name                      string
		Fields                    []FieldFacts
		Methods                   []MethodFacts
		ReferencedTypeKeys        []string
		Pos                       Position
		ExportedMembers           int
		DocumentedExportedMembers int
		Exported                  bool
		Kind                      TypeKind
	}

	// TypeKind classifies a named type's underlying form.
	TypeKind uint8

	// Position is a source location for a declaration.
	Position = factmodel.Position
)
