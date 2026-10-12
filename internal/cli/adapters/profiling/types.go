// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

// Package profiling writes CPU and heap profiles for CLI diagnostics.
package profiling

import (
	"os"
)

type (
	profileDeps = struct {
		writeHeap func(*os.File) error
		closeFile func(*os.File) error
	}
)
