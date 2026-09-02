// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

type (
	// ProjectFacts is the assembled, ID-assigned fact set for a module.
	ProjectFacts struct {
		// ModulePath is the import path of the main module, when known.
		ModulePath string
		// Packages is sorted by import path; a package's ID is its index.
		Packages []PackageFacts
		// Types is sorted by (package path, type name); a type's ID is its index.
		Types []TypeFacts
	}

	// PackageFacts is one analyzed package after ID assignment.
	PackageFacts struct {
		Path     string
		Imports  []string
		TypeIDs  []int
		ID       int
		InModule bool
	}

	// TypeFacts is one named type after ID assignment.
	TypeFacts struct {
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
