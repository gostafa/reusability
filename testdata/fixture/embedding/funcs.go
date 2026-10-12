// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package embedding

// Inc increments the counter.
func (b *Base) Inc() { b.Count++ }

// Describe touches the embedded slot and the own field.
func (w *Wrapper) Describe() string {
	w.Inc()
	w.Base.Count += 1
	if w.Count > 0 {
		return w.Name
	}
	return ""
}
