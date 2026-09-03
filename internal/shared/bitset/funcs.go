// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package bitset

import (
	"math/bits"
)

// Clone returns an independent copy of fieldSet.
func Clone(fieldSet FieldSet) FieldSet {
	return cloneOf(fieldSet)
}

// Clone returns an independent copy of the set.
func (fieldSet FieldSet) Clone() FieldSet {
	if fieldSet.words == nil {
		return FieldSet{}
	}

	words := make([]uint64, zero, len(fieldSet.words))

	words = append(words, fieldSet.words...)

	return FieldSet{words: words}
}

// Contains reports whether the field at index is set.
func Contains(fieldSet FieldSet, index int) bool {
	return containsAt(fieldSet, index)
}

// Contains reports whether the field at index is set.
func (fieldSet FieldSet) Contains(index int) bool {
	word := index / wordBits

	if word >= len(fieldSet.words) {
		return false
	}

	return fieldSet.words[word]&(one<<uint(index%wordBits)) != zero
}

// Count returns the number of set fields.
func Count(fieldSet FieldSet) int {
	return countOf(fieldSet)
}

// Count returns the number of set fields.
func (fieldSet FieldSet) Count() int {
	total := zero

	for i := range fieldSet.words {
		total += bits.OnesCount64(fieldSet.words[i])
	}

	return total
}

// Set marks the field at index. index must be within the set's capacity.
func (fieldSet FieldSet) Set(index int) {
	fieldSet.words[index/wordBits] |= one << uint(index%wordBits)
}

// Intersects reports whether the two sets share any field.
func Intersects(left, right FieldSet) bool {
	limit := min(len(left.words), len(right.words))

	for i := range limit {
		if left.words[i]&right.words[i] != zero {
			return true
		}
	}

	return false
}

// NewFieldSet returns a FieldSet able to hold size field indices.
func NewFieldSet(size int) FieldSet {
	if size <= zero {
		return FieldSet{}
	}

	return FieldSet{words: make([]uint64, (size+wordBits-one)/wordBits)}
}

// Small returns the single-word view of fieldSet. Valid only when the set was
// created for at most 64 fields.
func Small(fieldSet FieldSet) SmallFieldSet {
	if len(fieldSet.words) == zero {
		return SmallFieldSet{}
	}

	return SmallFieldSet{bits: fieldSet.words[zero]}
}

// Intersects reports whether the two sets share any field.
func (set SmallFieldSet) Intersects(other SmallFieldSet) bool {
	return set.bits&other.bits != zero
}

// Union adds every field of src to dst. src must not be wider than dst.
func Union(dst, src FieldSet) {
	for i := range src.words {
		dst.words[i] |= src.words[i]
	}
}

func cloneOf(source cloner) FieldSet {
	return source.Clone()
}

func containsAt(holder container, index int) bool {
	return holder.Contains(index)
}

func countOf(holder counter) int {
	return holder.Count()
}
