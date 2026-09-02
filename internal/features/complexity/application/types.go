// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"github.com/gostafa/reusability/internal/shared/metrics"
)

type (
	// Result is the complexity metrics for one named type.
	Result struct {
		MethodComplexities []int
		AMC                metrics.MetricResult
	}
)
