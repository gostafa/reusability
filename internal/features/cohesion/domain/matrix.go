package domain

import (
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/bitset"
)

// EffectiveFieldSets returns each method's field-usage set. In direct mode
// this is the extracted usage as-is; in transitive mode usage is propagated
// through calls to sibling methods until a fixpoint.
func EffectiveFieldSets(t *typefacts.TypeFacts, transitive bool) []bitset.FieldSet {
	sets := make([]bitset.FieldSet, len(t.Methods))
	for i := range t.Methods {
		sets[i] = t.Methods[i].FieldsUsed
	}

	if !transitive || len(t.Fields) == 0 {
		return sets
	}

	for i := range sets {
		sets[i] = bitset.Clone(sets[i])
	}

	for changed := true; changed; {
		changed = false

		for i := range t.Methods {
			for _, j := range t.Methods[i].CalledSiblings {
				before := bitset.Count(sets[i])
				bitset.Union(sets[i], sets[j])

				if bitset.Count(sets[i]) != before {
					changed = true
				}
			}
		}
	}

	return sets
}

// CountSharingPairs counts unordered method pairs whose field sets intersect.
// The O(k²) loop works on bitsets only; the single-word fast path applies
// whenever the type has at most 64 fields.
func CountSharingPairs(sets []bitset.FieldSet, fieldCount int) int {
	k := len(sets)

	var sharing int
	if k < 2 {
		return sharing
	}

	if fieldCount <= 64 {
		small := make([]bitset.SmallFieldSet, k)
		for i, s := range sets {
			small[i] = bitset.Small(s)
		}

		for i := range k {
			for j := i + 1; j < k; j++ {
				if small[i].Intersects(small[j]) {
					sharing++
				}
			}
		}

		return sharing
	}

	for i := range k {
		for j := i + 1; j < k; j++ {
			if bitset.Intersects(sets[i], sets[j]) {
				sharing++
			}
		}
	}

	return sharing
}

// TotalFieldAccesses is the number of 1-cells in the method-field matrix:
// each method contributes each distinct field it uses once.
func TotalFieldAccesses(sets []bitset.FieldSet) int {
	total := 0
	for _, s := range sets {
		total += bitset.Count(s)
	}

	return total
}
