// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package workerpool

import (
	"sync"
)

type (
	// RunConfig configures a bounded parallel map over task indices.
	RunConfig = struct {
		Fn        func(i int) error
		Workers   int
		TaskCount int
	}

	pool = struct {
		tasks chan int
		errs  []error
		wg    sync.WaitGroup
	}
)
