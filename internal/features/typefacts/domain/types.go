// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

// Package domain provides reusable analysis utilities.
package domain

import factmodel "github.com/gostafa/reusability/internal/features/typefacts/domain/model"

type (
	// FieldFacts describes one struct field.
	FieldFacts = factmodel.FieldFacts
	// DeclarationFacts is shared name/position/export metadata.
	DeclarationFacts = factmodel.DeclarationFacts
	// FunctionFacts describes a package-level function.
	FunctionFacts = factmodel.FunctionFacts
	// MethodFacts describes one method of a named type.
	MethodFacts = factmodel.MethodFacts
	// BranchStats counts control-flow constructs for cyclomatic complexity.
	BranchStats = factmodel.BranchStats
)

type (
	// PackagePathView exposes one package's import path.
	PackagePathView interface {
		ImportPath() string
	}

	// FactDumper renders a compact debug representation.
	FactDumper interface {
		String() string
	}

	// ProjectFacts is the assembled, ID-assigned fact set for a module.
	ProjectFacts = struct {
		// ModulePath is the import path of the main module, when known.
		ModulePath string
		// Packages is sorted by import path; a package's ID is its index.
		Packages []PackageFacts
		// Types is sorted by (package path, type name); a type's ID is its index.
		Types []TypeFacts
	}

	// PackageFacts is one analyzed package after ID assignment.
	PackageFacts = struct {
		Path     string
		Imports  []string
		TypeIDs  []int
		ID       int
		InModule bool
	}

	// TypeFacts is one named type after ID assignment.
	TypeFacts = struct {
		Name                      string
		Fields                    []FieldFacts
		Methods                   []MethodFacts
		ReferencedTypeIDs         []int
		Pos                       Position
		ID                        int
		PackageID                 int
		ExportedMembers           int
		DocumentedExportedMembers int
		Exported                  bool
		Kind                      TypeKind
	}
)

type (
	// kindViewer exposes a named type's underlying kind code.
	kindViewer interface {
		Code() uint8
	}

	// PackageExtract is one package's raw extraction before ID assignment.
	PackageExtract = struct {
		// Path is the package import path.
		Path string
		// Imports lists import paths referenced by the package.
		Imports []string
		// Types holds named types extracted from the package.
		Types []TypeExtract
		// InModule reports whether the package belongs to the main module.
		InModule bool
	}

	// TypeExtract is one named type's raw extraction before ID assignment.
	TypeExtract = struct {
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
