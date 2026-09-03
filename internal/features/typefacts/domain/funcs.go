// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"fmt"
	"strconv"
)

// Code returns the underlying kind discriminant.
func (kind TypeKind) Code() uint8 {
	return uint8(kind)
}

// String returns the kind discriminant as a decimal string.
func (kind TypeKind) String() string {
	return strconv.FormatUint(uint64(kindCode(kind)), decimalBase)
}

func projectFactsString(facts *ProjectFacts) string {
	return fmt.Sprintf(
		"module %q: %d packages, %d types",
		facts.ModulePath,
		len(facts.Packages),
		len(facts.Types),
	)
}

// TypeKey is the canonical cross-package key of a named type.
func TypeKey(pkgPath, typeName string) string {
	return pkgPath + "." + typeName
}

func kindCode(view kindViewer) uint8 {
	return view.Code()
}

func typeExtractString(extract *TypeExtract) string {
	return fmt.Sprintf(
		"type %q (kind %d, exported %v) at %v: %d fields, %d methods, %d refs, %d/%d documented",
		extract.Name,
		kindCode(extract.Kind),
		extract.Exported,
		extract.Pos,
		len(extract.Fields),
		len(extract.Methods),
		len(extract.ReferencedTypeKeys),
		extract.DocumentedExportedMembers,
		extract.ExportedMembers,
	)
}

func typeFactsString(facts *TypeFacts) string {
	return fmt.Sprintf(
		"type %d %q (pkg %d, kind %d, exported %v) at %v: %d fields, %d methods, %d refs, %d/%d docs",
		facts.ID,
		facts.Name,
		facts.PackageID,
		kindCode(facts.Kind),
		facts.Exported,
		facts.Pos,
		len(facts.Fields),
		len(facts.Methods),
		len(facts.ReferencedTypeIDs),
		facts.DocumentedExportedMembers,
		facts.ExportedMembers,
	)
}
