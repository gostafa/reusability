// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package bitset

type (
	// SmallFieldSet is a single-word field-usage bitset (≤64 fields).
	SmallFieldSet struct {
		bits uint64
	}

	// FieldSet is a multi-word field-usage bitset.
	FieldSet struct {
		words []uint64
	}

	// Setter marks a field index as used.
	Setter interface {
		Set(index int)
	}

	// Intersecter reports whether two small field sets share a field.
	Intersecter interface {
		Intersects(other SmallFieldSet) bool
	}

	container interface {
		Contains(index int) bool
	}

	counter interface {
		Count() int
	}

	cloner interface {
		Clone() FieldSet
	}
)
