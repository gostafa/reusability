// Package embedding exercises embedded and promoted members.
package embedding

// Base counts things.
type Base struct {
	Count int
}

// Inc increments the counter.
func (b *Base) Inc() { b.Count++ }

// Wrapper embeds Base; promoted members stay Base's.
type Wrapper struct {
	Base
	Name string
}

// Describe touches the embedded slot and the own field.
func (w *Wrapper) Describe() string {
	w.Inc()           // promoted method: not a Wrapper method
	w.Base.Count += 1 // uses Wrapper's Base slot; Count stays Base's field
	if w.Count > 0 {  // promoted field: belongs to Base, not Wrapper
		return w.Name
	}
	return ""
}
