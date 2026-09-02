// Package multifile spreads methods across files.
package multifile

// Counter tallies hits and misses.
type Counter struct {
	hits   int
	misses int
}

// Hit records a hit.
func (c *Counter) Hit() { c.hits++ }
