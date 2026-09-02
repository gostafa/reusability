// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"fmt"
)

// String summarizes the fact set for debugging.
func (f *ProjectFacts) String() string {
	return fmt.Sprintf(
		"module %q: %d packages, %d types",
		f.ModulePath,
		len(f.Packages),
		len(f.Types),
	)
}

// String summarizes the extract for debugging.
func typeExtractString(t *TypeExtract) string {
	return fmt.Sprintf(
		"type %q (kind %d, exported %v) at %v: %d fields, %d methods, %d refs, %d/%d documented",
		t.Name,
		t.Kind,
		t.Exported,
		t.Pos,
		len(t.Fields),
		len(t.Methods),
		len(t.ReferencedTypeKeys),
		t.DocumentedExportedMembers,
		t.ExportedMembers,
	)
}

// String summarizes the type facts for debugging.
func typeFactsString(t *TypeFacts) string {
	return fmt.Sprintf(
		"type %d %q (pkg %d, kind %d, exported %v) at %v: %d fields, %d methods, %d refs, %d/%d docs",
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

// TypeKey is the canonical cross-package key of a named type.
func TypeKey(pkgPath, typeName string) string {
	return pkgPath + "." + typeName
}
