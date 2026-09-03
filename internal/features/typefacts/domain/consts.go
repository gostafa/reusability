// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

const (
	// KindStruct marks a named type whose underlying type is a struct.
	KindStruct TypeKind = 0
	// KindInterface marks a named type whose underlying type is an interface.
	KindInterface TypeKind = 1
	// KindOther marks any other named type (basic, slice, func, …).
	KindOther TypeKind = 2

	decimalBase = 10
)
