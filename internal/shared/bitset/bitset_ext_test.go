package bitset_test

import (
	"testing"

	"github.com/gostafa/reusability/internal/shared/bitset"
)

// Black-box: the exported multi-word set operations.
func TestFieldSetOperations(t *testing.T) {
	t.Parallel()

	a := bitset.NewFieldSet(80) // spans two words
	a.Set(0)
	a.Set(65)

	if !bitset.Contains(a, 0) || !bitset.Contains(a, 65) {
		t.Fatal("Set/Contains mismatch")
	}

	if bitset.Contains(a, 1) {
		t.Fatal("unset bit reported present")
	}

	if bitset.Count(a) != 2 {
		t.Fatalf("Count = %d, want 2", bitset.Count(a))
	}

	b := bitset.NewFieldSet(80)
	b.Set(1)
	bitset.Union(a, b)

	if bitset.Count(a) != 3 {
		t.Fatalf("Union count = %d, want 3", bitset.Count(a))
	}

	if !bitset.Intersects(a, b) {
		t.Fatal("a should intersect b on bit 1")
	}

	c := bitset.Clone(a)
	c.Set(2)

	if bitset.Count(c) == bitset.Count(a) {
		t.Fatal("Clone must be independent of the original")
	}
}

// Black-box: the single-word fast path.
func TestSmallFieldSet(t *testing.T) {
	t.Parallel()

	a := bitset.NewFieldSet(4)
	a.Set(0)

	b := bitset.NewFieldSet(4)
	b.Set(0)

	if !bitset.Small(a).Intersects(bitset.Small(b)) {
		t.Fatal("small sets share bit 0")
	}

	d := bitset.NewFieldSet(4)
	d.Set(3)

	if bitset.Small(a).Intersects(bitset.Small(d)) {
		t.Fatal("disjoint small sets must not intersect")
	}
}
