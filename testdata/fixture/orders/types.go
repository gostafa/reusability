// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

// Package orders holds the order domain.
package orders

// Order is a purchase order.
type Order struct {
	// ID identifies the order.
	ID    int
	Total float64
	notes string
}
