// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

// Package generics exercises type parameters.
package generics

// Pair holds two values of the same type.
type Pair[T any] struct {
	First  T
	Second T
}
