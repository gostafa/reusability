// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/bitset"
)

// CountSharingPairs counts unordered method pairs whose field sets intersect.
// The O(k²) loop works on bitsets only; the single-word fast path applies
// whenever the type has at most 64 fields.
func CountSharingPairs(sets []bitset.FieldSet, fieldCount int) int {
	if len(sets) < pairMinMethods {
		return zero
	}

	if fieldCount <= smallFieldCap {
		return countSmallPairs(toSmallSets(sets))
	}

	return countWidePairs(sets)
}

// EffectiveFieldSets returns each method's field-usage set. In direct mode
// this is the extracted usage as-is; in transitive mode usage is propagated
// through calls to sibling methods until a fixpoint.
func EffectiveFieldSets(t *typefacts.TypeFacts, mode FieldUsageMode) []bitset.FieldSet {
	sets := directFieldSets(t)

	if mode != FieldUsageTransitive || len(t.Fields) == zero {
		return sets
	}

	return propagateFieldSets(t, cloneSets(sets))
}

// TotalFieldAccesses is the number of 1-cells in the method-field matrix:
// each method contributes each distinct field it uses once.
func TotalFieldAccesses(sets []bitset.FieldSet) int {
	total := zero

	for index := range sets {
		total += bitset.Count(sets[index])
	}

	return total
}

func cloneSets(sets []bitset.FieldSet) []bitset.FieldSet {
	out := make([]bitset.FieldSet, zero, len(sets))

	for index := range sets {
		out = append(out, bitset.Clone(sets[index]))
	}

	return out
}

func countSmallPairs(small []bitset.SmallFieldSet) int {
	sharing := zero

	for index := range small {
		sharing += countSmallFrom(small, index)
	}

	return sharing
}

func countSmallFrom(small []bitset.SmallFieldSet, index int) int {
	sharing := zero

	for other := index + one; other < len(small); other++ {
		if small[index].Intersects(small[other]) {
			sharing++
		}
	}

	return sharing
}

func countWidePairs(sets []bitset.FieldSet) int {
	sharing := zero

	for index := range sets {
		sharing += countWideFrom(sets, index)
	}

	return sharing
}

func countWideFrom(sets []bitset.FieldSet, index int) int {
	sharing := zero

	for other := index + one; other < len(sets); other++ {
		if bitset.Intersects(sets[index], sets[other]) {
			sharing++
		}
	}

	return sharing
}

func directFieldSets(t *typefacts.TypeFacts) []bitset.FieldSet {
	sets := make([]bitset.FieldSet, zero, len(t.Methods))

	for index := range t.Methods {
		sets = append(sets, t.Methods[index].FieldsUsed)
	}

	return sets
}

func propagateFieldSets(t *typefacts.TypeFacts, sets []bitset.FieldSet) []bitset.FieldSet {
	for changed := true; changed; {
		changed = propagateRound(t, sets)
	}

	return sets
}

func propagateRound(t *typefacts.TypeFacts, sets []bitset.FieldSet) bool {
	changed := false

	for index := range t.Methods {
		if absorbSiblings(sets, index, t.Methods[index].CalledSiblings) {
			changed = true
		}
	}

	return changed
}

func absorbSiblings(sets []bitset.FieldSet, index int, siblings []int) bool {
	changed := false

	for sibling := range siblings {
		before := bitset.Count(sets[index])
		bitset.Union(sets[index], sets[siblings[sibling]])

		if bitset.Count(sets[index]) != before {
			changed = true
		}
	}

	return changed
}

func toSmallSets(sets []bitset.FieldSet) []bitset.SmallFieldSet {
	small := make([]bitset.SmallFieldSet, zero, len(sets))

	for index := range sets {
		small = append(small, bitset.Small(sets[index]))
	}

	return small
}
