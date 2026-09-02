// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package bitset

import "testing"

// Black-box: the exported multi-word set operations.
func TestFieldSetOperations(t *testing.T) {
	t.Parallel()

	a := NewFieldSet(80)
	a.Set(0)
	a.Set(65)

	if !Contains(a, 0) || !Contains(a, 65) {
		t.Fatal("Set/Contains mismatch")
	}

	if Contains(a, 1) {
		t.Fatal("unset bit reported present")
	}

	if Count(a) != 2 {
		t.Fatalf("Count = %d, want 2", Count(a))
	}

	b := NewFieldSet(80)
	b.Set(1)
	Union(a, b)

	if Count(a) != 3 {
		t.Fatalf("Union count = %d, want 3", Count(a))
	}

	if !Intersects(a, b) {
		t.Fatal("a should intersect b on bit 1")
	}

	c := Clone(a)
	c.Set(2)

	if Count(c) == Count(a) {
		t.Fatal("Clone must be independent of the original")
	}
}

// Black-box: the single-word fast path.
func TestSmallFieldSet(t *testing.T) {
	t.Parallel()

	a := NewFieldSet(4)
	a.Set(0)

	b := NewFieldSet(4)
	b.Set(0)

	if !Small(a).Intersects(Small(b)) {
		t.Fatal("small sets share bit 0")
	}

	d := NewFieldSet(4)
	d.Set(3)

	if Small(a).Intersects(Small(d)) {
		t.Fatal("disjoint small sets must not intersect")
	}
}
