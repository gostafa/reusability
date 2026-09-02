// Package orders holds the order domain.
package orders

import (
	"fmt"

	"example.com/fixture/store"
)

// Order is a purchase order.
type Order struct {
	// ID identifies the order.
	ID    int
	Total float64
	notes string
}

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
