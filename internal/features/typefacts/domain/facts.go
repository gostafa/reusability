package domain

import (
	"fmt"

	factmodel "github.com/gostafa/reusability/internal/features/typefacts/domain/model"
)

// TypeKind classifies a named type's underlying type.
type TypeKind uint8

const (
	// KindStruct marks a named type whose underlying type is a struct.
	KindStruct TypeKind = iota
	// KindInterface marks a named type whose underlying type is an interface.
	KindInterface
	// KindOther marks any other named type (basic, slice, func, …).
	KindOther
)

// Position locates a declaration in source, with File relative to the
// analysis directory when possible so output is machine-independent.
type Position = factmodel.Position

// ProjectFacts is everything the metric features need to know about the
// analyzed packages. All slices are deterministically ordered and all
// cross-references use dense numeric IDs (indices into the slices).
type ProjectFacts struct {
	// ModulePath is the import path of the main module, when known.
	ModulePath string
	// Packages is sorted by import path; a package's ID is its index.
	Packages []PackageFacts
	// Types is sorted by (package path, type name); a type's ID is its index.
	Types []TypeFacts
}

// String summarizes the fact set for debugging.
func (f *ProjectFacts) String() string {
	return fmt.Sprintf(
		"module %q: %d packages, %d types",
		f.ModulePath,
		len(f.Packages),
		len(f.Types),
	)
}

// PackageFacts describes one analyzed package.
type PackageFacts struct {
	// ID is the package's dense index into ProjectFacts.Packages.
	ID int
	// Path is the package's import path.
	Path string
	// InModule reports whether the package belongs to the main module.
	InModule bool
	// Imports are the package's distinct import paths, sorted, without
	// self-imports. Scope filtering happens in the architecture feature.
	Imports []string
	// TypeIDs are the package's analyzed types in name order.
	TypeIDs []int
}

// TypeFacts describes one analyzed named type. Aliases are never analyzed.
type TypeFacts struct {
	// ID is the type's dense index into ProjectFacts.Types.
	ID int
	// PackageID is the declaring package's dense index.
	PackageID int
	// Name is the type's declared name.
	Name string
	// Exported reports whether the type name is exported.
	Exported bool
	// Kind classifies the type's underlying type.
	Kind TypeKind
	// Pos locates the type declaration in source.
	Pos Position
	// Fields are the struct's fields in declaration order; empty for
	// non-struct types. An embedded field occupies exactly one slot and
	// promoted members are not represented (they belong to the embedded
	// type).
	Fields []FieldFacts
	// Methods are the explicitly declared methods (functions with a
	// receiver, pointer and value receivers normalized to this type),
	// sorted by name then source position. Promoted methods are excluded.
	Methods []MethodFacts
	// ReferencedTypeIDs is the CBO fact: the distinct other analyzed types
	// this type references through fields, method parameters, method
	// returns, and embedded types. Sorted, self excluded.
	ReferencedTypeIDs []int
	// ExportedMembers counts the type's exported members (the type itself
	// when exported, plus exported declared methods and exported fields).
	ExportedMembers int
	// DocumentedExportedMembers counts exported members carrying a doc
	// comment.
	DocumentedExportedMembers int
}

// String summarizes the type facts for debugging.
func (t *TypeFacts) String() string {
	return fmt.Sprintf(
		"type %d %q (package %d, kind %d, exported %v) at %v: %d fields, %d methods, %d refs, %d/%d documented exported members",
		t.ID,
		t.Name,
		t.PackageID,
		t.Kind,
		t.Exported,
		t.Pos,
		len(t.Fields),
		len(t.Methods),
		len(t.ReferencedTypeIDs),
		t.DocumentedExportedMembers,
		t.ExportedMembers,
	)
}

// FieldFacts describes one struct field slot.
type FieldFacts = factmodel.FieldFacts

// DeclarationFacts describes one top-level named var or const declaration.
type DeclarationFacts = factmodel.DeclarationFacts

// FunctionFacts describes one top-level function declaration.
type FunctionFacts = factmodel.FunctionFacts

// MethodFacts describes one explicitly declared method.
type MethodFacts = factmodel.MethodFacts

// BranchStats counts the syntax constructs that increment cyclomatic
// complexity. The formula itself lives in the complexity feature.
type BranchStats = factmodel.BranchStats
