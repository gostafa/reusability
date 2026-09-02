// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"github.com/gostafa/reusability/internal/shared/metrics"
)

type (
	// Result is the cohesion metrics for one named type.
	Result struct {
		// LCOM is the method-field matrix sparsity.
		LCOM metrics.MetricResult
		// TCC is the connected method-pair density.
		TCC metrics.MetricResult
	}
)
