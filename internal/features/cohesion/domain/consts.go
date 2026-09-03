// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

const (
	// FieldUsageDirect counts only fields a method body accesses directly.
	FieldUsageDirect FieldUsageMode = "direct"
	// FieldUsageTransitive also propagates usage through sibling calls.
	FieldUsageTransitive FieldUsageMode = "transitive"

	zero           = 0
	one            = 1
	pairMinMethods = 2
	smallFieldCap  = 64
)
