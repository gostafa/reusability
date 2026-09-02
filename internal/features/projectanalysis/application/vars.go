// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"github.com/gostafa/reusability/internal/features/projectanalysis/ports/inbound"
	"github.com/gostafa/reusability/internal/shared/workerpool"
)

var (
	_ inbound.Analyzer = (*Pipeline)(nil)

	// runWorkers is the pool entry point; tests may swap it.
	runWorkers = workerpool.Run
)
