package multifile

// Total sums both tallies.
func (c Counter) Total() int { return c.total() }

func (c Counter) total() int { return c.hits + c.misses }
