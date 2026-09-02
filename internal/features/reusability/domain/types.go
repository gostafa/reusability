// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"github.com/gostafa/reusability/internal/shared/metrics"
)

type (
	// Result holds CBO and the reusability index for one type.
	Result struct {
		// CBO is the normalized coupling input, reported standalone only when
		// selected.
		CBO metrics.MetricResult
		// Reusability is the experimental reusability index.
		Reusability metrics.MetricResult
	}
)
