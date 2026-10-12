// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package orders

import (
	"fmt"

	"example.com/fixture/store"
)

// GrandTotal applies the tax rate when both amounts are positive.
func (o *Order) GrandTotal(tax float64) float64 {
	if tax > 0 && o.Total > 0 {
		return o.Total * (1 + tax)
	}
	return o.Total
}

func (o Order) Note() string { return fmt.Sprintf("note: %s", o.notes) }

// Save persists the order.
func (o *Order) Save(s store.Store) error { return s.Put(o.ID, o) }
