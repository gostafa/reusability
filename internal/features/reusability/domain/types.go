// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
)

type (
	// ComputeInput bundles the facts needed to evaluate one type's index.
	ComputeInput struct {
		Type    *typefacts.TypeFacts
		AMC     metrics.MetricResult
		LCOM    metrics.MetricResult
		Weights metrics.ReusabilityWeights
	}

	// Result holds CBO and the reusability index for one type.
	Result struct {
		// CBO is the normalized coupling input, reported standalone only when
		// selected.
		CBO metrics.MetricResult
		// Reusability is the experimental reusability index.
		Reusability metrics.MetricResult
	}
)
