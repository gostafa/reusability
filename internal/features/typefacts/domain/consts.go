// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

const (

	// KindStruct marks a named type whose underlying type is a struct.
	KindStruct TypeKind = iota
	// KindInterface marks a named type whose underlying type is an interface.
	KindInterface
	// KindOther marks any other named type (basic, slice, func, …).
	KindOther
)
