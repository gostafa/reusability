package bitset

import "testing"

func TestSmallFieldSet(t *testing.T) {
	a := NewFieldSet(64)
	a.Set(0)
	a.Set(63)

	b := NewFieldSet(64)
	b.Set(1)

	if Small(a).Intersects(Small(b)) {
		t.Fatal("disjoint small sets intersect")
	}

	b.Set(63)

	if !Small(a).Intersects(Small(b)) {
		t.Fatal("overlapping small sets do not intersect")
	}
}

func TestFieldSetGeneralPath(t *testing.T) {
	// 65 fields exercises the multi-word path across the 64-bit boundary.
	a := NewFieldSet(65)
	b := NewFieldSet(65)

	a.Set(0)
	a.Set(64)

	if !Contains(a, 64) || Contains(a, 63) {
		t.Fatal("contains across word boundary")
	}

	if Count(a) != 2 {
		t.Fatalf("count = %d, want 2", Count(a))
	}

	b.Set(63)

	if Intersects(a, b) {
		t.Fatal("disjoint sets intersect")
	}

	b.Set(64)

	if !Intersects(a, b) {
		t.Fatal("overlapping sets do not intersect")
	}

	Union(a, b)

	if Count(a) != 3 {
		t.Fatalf("union count = %d, want 3", Count(a))
	}

	clone := Clone(a)
	clone.Set(10)

	if Contains(a, 10) {
		t.Fatal("clone shares storage with original")
	}
}

func TestFieldSetSmallView(t *testing.T) {
	a := NewFieldSet(8)
	a.Set(3)

	probe := NewFieldSet(8)
	probe.Set(3)

	if !Small(a).Intersects(Small(probe)) {
		t.Fatal("small view lost the set field")
	}

	if Count(NewFieldSet(0)) != 0 {
		t.Fatal("zero-size set should be empty")
	}

	var nilSet FieldSet
	if Count(nilSet) != 0 || Contains(nilSet, 5) || Small(nilSet).Intersects(Small(probe)) {
		t.Fatal("empty set operations")
	}

	if Intersects(nilSet, a) {
		t.Fatal("nil set intersects")
	}

	if clone := Clone(FieldSet{}); Count(clone) != 0 || Contains(clone, 0) {
		t.Fatal("Clone of empty FieldSet must stay empty")
	}
}
