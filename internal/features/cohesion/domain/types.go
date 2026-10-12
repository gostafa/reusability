// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

// Package domain provides reusable analysis utilities.
package domain

type (
	// modeStringer exposes a field-usage mode's canonical name.
	modeStringer interface {
		String() string
	}

	// FieldUsageMode selects whether field usage is direct or call-transitive.
	FieldUsageMode string
)
