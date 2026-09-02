package bitset

import "math/bits"

// SmallFieldSet is a single-word bitset for up to 64 field indices.
type SmallFieldSet struct {
	bits uint64
}

// Intersects reports whether the two sets share any field.
func (s SmallFieldSet) Intersects(other SmallFieldSet) bool {
	return s.bits&other.bits != 0
}

const wordBits = 64

// FieldSet is a multi-word bitset for any number of field indices. The
// zero value is an empty set.
type FieldSet struct {
	words []uint64
}

// NewFieldSet returns a FieldSet able to hold size field indices.
func NewFieldSet(size int) FieldSet {
	if size <= 0 {
		return FieldSet{}
	}

	return FieldSet{words: make([]uint64, (size+wordBits-1)/wordBits)}
}

// Set marks the field at index. index must be within the set's capacity.
func (f FieldSet) Set(index int) {
	f.words[index/wordBits] |= 1 << uint(index%wordBits)
}

// Contains reports whether the field at index is set.
func Contains(f FieldSet, index int) bool {
	word := index / wordBits
	if word >= len(f.words) {
		return false
	}

	return f.words[word]&(1<<uint(index%wordBits)) != 0
}

// Count returns the number of set fields.
func Count(f FieldSet) int {
	total := 0
	for _, word := range f.words {
		total += bits.OnesCount64(word)
	}

	return total
}

// Union adds every field of src to dst. src must not be wider than dst.
func Union(dst, src FieldSet) {
	for i := range src.words {
		dst.words[i] |= src.words[i]
	}
}

// Intersects reports whether the two sets share any field.
func Intersects(a, b FieldSet) bool {
	n := min(len(a.words), len(b.words))
	for i := range n {
		if a.words[i]&b.words[i] != 0 {
			return true
		}
	}

	return false
}

// Clone returns an independent copy of f.
func Clone(f FieldSet) FieldSet {
	if f.words == nil {
		return FieldSet{}
	}

	words := make([]uint64, len(f.words))
	copy(words, f.words)

	return FieldSet{words: words}
}

// Small returns the single-word view of f. Valid only when the set was
// created for at most 64 fields.
func Small(f FieldSet) SmallFieldSet {
	if len(f.words) == 0 {
		return SmallFieldSet{}
	}

	return SmallFieldSet{bits: f.words[0]}
}

// Setter is the write side of a field-usage set: it records that the field at
// the given index is used. FieldSet implements it.
type Setter interface {
	Set(index int)
}

// Intersecter reports whether it shares any set bit with another single-word
// set. SmallFieldSet implements it.
type Intersecter interface {
	Intersects(other SmallFieldSet) bool
}

var (
	_ Setter      = FieldSet{}
	_ Intersecter = SmallFieldSet{}
)
