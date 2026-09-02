// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"github.com/gostafa/reusability/internal/shared/metrics"
)

type (
	// Result is the complexity metrics for one named type.
	Result struct {
		// MethodComplexities are per-method cyclomatic complexities in
		// declaration order.
		MethodComplexities []int
		// AMC is the average method complexity for the type.
		AMC metrics.MetricResult
	}
)
