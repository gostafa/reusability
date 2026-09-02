// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"fmt"

	"github.com/gostafa/reusability/internal/features/reusability/domain"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
)

// NewService validates the weights and returns a reusability evaluator.
// Zero-value weights select the defaults.
func NewService(weights *metrics.ReusabilityWeights) (*Service, error) {
	resolved := metrics.DefaultReusabilityWeights()

	if weights != nil && *weights != (metrics.ReusabilityWeights{}) {
		resolved = *weights
	}

	err := resolved.Validate()
	if err != nil {
		return nil, fmt.Errorf("NewService: %w", err)
	}

	return &Service{weights: resolved}, nil
}

// ComputeForType evaluates CBO and the reusability index for one type.
// fieldUsage is "direct" or "transitive".
func (svc *Service) ComputeForType(typeFacts *typefacts.TypeFacts, fieldUsage string) Result {
	return domain.Compute(typeFacts, svc.weights, fieldUsage)
}
