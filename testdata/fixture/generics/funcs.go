// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package generics

// SetFirst stores v in the first slot.
func (p *Pair[T]) SetFirst(v T) { p.First = v }

// SetSecond stores v in the second slot.
func (p *Pair[T]) SetSecond(v T) { p.Second = v }

// Swapped returns a pair with other's slots exchanged.
func (p Pair[T]) Swapped(other Pair[T]) Pair[T] {
	return Pair[T]{First: other.Second, Second: other.First}
}
