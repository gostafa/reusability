// Package isolated depends on nothing inside the module.
package isolated

type Value float64

// Double doubles the value.
func (v Value) Double() Value { return v * 2 }
