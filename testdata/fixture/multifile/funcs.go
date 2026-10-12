// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package multifile

// Hit records a hit.
func (c *Counter) Hit() { c.hits++ }

// Total sums both tallies.
func (c Counter) Total() int { return c.total() }

func (c Counter) total() int { return c.hits + c.misses }
