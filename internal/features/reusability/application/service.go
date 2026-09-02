package application

import (
	"github.com/gostafa/reusability/internal/features/reusability/domain"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
)

// Result re-exports the domain result for consumers of the feature.
type Result = domain.Result

// Calculator is the reusability application boundary consumed by analysis
// pipelines. Implementations evaluate CBO and reusability for one type.
type Calculator interface {
	ComputeForType(*typefacts.TypeFacts, metrics.MetricResult, metrics.MetricResult) Result
}

// Service evaluates the reusability index with a fixed weight set.
type Service struct {
	weights metrics.ReusabilityWeights
}

var _ Calculator = (*Service)(nil)

// NewService validates the weights and returns a reusability evaluator.
// Zero-value weights select the defaults.
func NewService(weights metrics.ReusabilityWeights) (*Service, error) {
	if (weights == metrics.ReusabilityWeights{}) {
		weights = metrics.DefaultReusabilityWeights()
	}

	err := weights.Validate()
	if err != nil {
		return nil, err
	}

	return &Service{weights: weights}, nil
}

// ComputeForType evaluates CBO and the reusability index for one type. The
// AMC and LCOM results are supplied by the orchestrator.
func (s *Service) ComputeForType(t *typefacts.TypeFacts, amc, lcom metrics.MetricResult) Result {
	return domain.Compute(t, amc, lcom, s.weights)
}
