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

// ComputeForType evaluates CBO and the reusability index for one type. The
// AMC and LCOM results are supplied by the orchestrator.
func (svc *Service) ComputeForType(
	typeFacts *typefacts.TypeFacts,
	amc, lcom *metrics.MetricResult,
) Result {
	// Domain compute assembles CBO plus the weighted reusability index.
	return domain.Compute(&domain.ComputeInput{
		Type: typeFacts, AMC: *amc, LCOM: *lcom, Weights: svc.weights,
	})
}
